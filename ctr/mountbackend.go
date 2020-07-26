package ctr

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/containerd/containerd/mount"
	"github.com/hashicorp/go-multierror"
	oci "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opentracing/opentracing-go/log"
)

type MountBackend interface {
	// TODO make MountTree fields accessible outside this pkg so external impls
	// of MountBackend can be written
	SetupTree(*MountTree, ContainerState) (mounts []oci.Mount, diffDirs map[string]string, cleanup CleanupStack, err error)
}

type OverlayBackend struct{}

func (OverlayBackend) SetupTree(mt *MountTree, cs ContainerState) (_ []oci.Mount, _ map[string]string, cleanup CleanupStack, rerr error) {
	defer func() {
		if rerr != nil && len(cleanup) > 0 {
			rerr = multierror.Append(rerr, cleanup.Cleanup()).ErrorOrNil()
		}
	}()

	var ocimounts []oci.Mount
	diffDirs := make(map[string]string)
	var lowerdirIndex uint
	curTrees := []*tree{mt.tree}
	for len(curTrees) > 0 {
		var nextTrees []*tree
		for _, curTree := range curTrees {
			nextTrees = append(nextTrees, curTree.submounts...)
			if curTree.ociMount != nil {
				ocimounts = append(ocimounts, oci.Mount(*curTree.ociMount))
				continue
			}
			if len(curTree.srcs) == 0 {
				continue
			}

			overlayDir := cs.OverlayDir(curTree.mountpoint)
			upperDir := overlayDir.UpperDir()
			workDir := overlayDir.WorkDir()
			privateDir := overlayDir.PrivateDir()

			if err := os.MkdirAll(upperDir, 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf("failed to create upper dir: %w", err)
			}

			if err := os.MkdirAll(workDir, 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf("failed to create work dir: %w", err)
			}
			cleanup = append(cleanup, func() error {
				return os.RemoveAll(workDir)
			})

			if err := os.MkdirAll(overlayDir.PrivateDir(), 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf(
					"failed to create private lower dir: %w", err)
			}
			cleanup = append(cleanup, func() error {
				return os.RemoveAll(privateDir)
			})

			// setup private lower dir, which masks the dests on top of this one from showing
			// up in the upperdir changes
			for _, nextTree := range nextTrees {
				if err := setupPrivateDir(curTree, nextTree, privateDir); err != nil {
					return nil, nil, cleanup, err
				}
			}

			// setup shorthand lowerdir symlinks, which
			// help keep the length of the options provided to the mount syscall
			// under its 1 page size limit
			lowerdirs := make([]string, len(curTree.srcs)+1)
			for i, isrc := range append([]indexedSrc{{src: privateDir}}, curTree.srcs...) {
				lowerdir := cs.lowerDirSymlink(lowerdirIndex)
				lowerdirIndex += 1
				// overlay has lowerdirs from top->bottom, but we store them by
				// appending top layers to the end (bottom->top), so reverse srcs
				lowerdirs[len(curTree.srcs)-i] = filepath.Base(lowerdir)

				if err := os.Symlink(isrc.src, lowerdir); err != nil && !os.IsNotExist(err) {
					return nil, nil, cleanup, fmt.Errorf("failed to symlink lowerdir: %w", err)
				}
			}

			ocimounts = append(ocimounts, oci.Mount{
				Source:      "none",
				Destination: curTree.mountpoint,
				Type:        "overlay",
				Options: overlayOptions{
					LowerDirs: lowerdirs,
					UpperDir:  upperDir,
					WorkDir:   workDir,
				}.OptionsSlice(),
			})
			diffDirs[curTree.mountpoint] = upperDir
		}
		curTrees = nextTrees
	}
	return ocimounts, diffDirs, cleanup, nil
}

type FuseOverlayfsBackend struct {
	FuseOverlayfsBin string
}

func (b FuseOverlayfsBackend) SetupTree(mt *MountTree, cs ContainerState) (_ []oci.Mount, _ map[string]string, cleanup CleanupStack, rerr error) {
	defer func() {
		if rerr != nil && len(cleanup) > 0 {
			rerr = multierror.Append(rerr, cleanup.Cleanup()).ErrorOrNil()
		}
	}()

	var ocimounts []oci.Mount
	diffDirs := make(map[string]string)
	curTrees := []*tree{mt.tree}
	for len(curTrees) > 0 {
		var nextTrees []*tree
		for _, curTree := range curTrees {
			nextTrees = append(nextTrees, curTree.submounts...)
			if curTree.ociMount != nil {
				ocimounts = append(ocimounts, oci.Mount(*curTree.ociMount))
				continue
			}
			if len(curTree.srcs) == 0 {
				continue
			}

			overlayDir := cs.OverlayDir(curTree.mountpoint)
			upperDir := overlayDir.UpperDir()
			workDir := overlayDir.WorkDir()
			privateDir := overlayDir.PrivateDir()

			if err := os.MkdirAll(upperDir, 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf("failed to create upper dir: %w", err)
			}

			if err := os.MkdirAll(workDir, 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf("failed to create work dir: %w", err)
			}
			cleanup = append(cleanup, func() error {
				return os.RemoveAll(workDir)
			})

			if err := os.MkdirAll(overlayDir.PrivateDir(), 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf(
					"failed to create private lower dir: %w", err)
			}
			cleanup = append(cleanup, func() error {
				return os.RemoveAll(privateDir)
			})

			// setup private lower dir, which masks the dests on top of this one from showing
			// up in the upperdir changes
			for _, nextTree := range nextTrees {
				if err := setupPrivateDir(curTree, nextTree, privateDir); err != nil {
					return nil, nil, cleanup, err
				}
			}

			lowerdirs := make([]string, len(curTree.srcs)+1)
			for i, isrc := range append([]indexedSrc{{src: privateDir}}, curTree.srcs...) {
				// overlay has lowerdirs from top->bottom, but we store them by
				// appending top layers to the end (bottom->top), so reverse srcs
				lowerdirs[len(curTree.srcs)-i] = isrc.src
			}

			mountDir := overlayDir.MountDir()
			if err := os.MkdirAll(mountDir, 0700); err != nil {
				return nil, nil, cleanup, fmt.Errorf("failed to create overlay mount dir: %w", err)
			}

			cmd := exec.Cmd{
				Path: b.FuseOverlayfsBin,
				Args: []string{b.FuseOverlayfsBin,
					"-f", // stay in foreground
					"-o", overlayOptions{
						LowerDirs: lowerdirs,
						UpperDir:  upperDir,
						WorkDir:   workDir,
						Extra:     []string{"clone_fd"},
					}.Options(),
					mountDir,
				},
				Dir: cs.rootfsDir(),
				// TODO optional way to enable debug mode and send output somewhere?
				// Stdout: os.Stdout,
				// Stderr: os.Stderr,
			}

			timeoutCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
			defer done()
			if err := waitForMountChangeAt(timeoutCtx, mountDir, func(context.Context) error {
				if err := cmd.Start(); err != nil {
					return fmt.Errorf("failed to start fuse-overlayfs for mount dir %s: %w", mountDir, err)
				}

				waitCh := make(chan error)
				go func() {
					defer close(waitCh)
					waitCh <- cmd.Wait()
				}()

				cleanup = cleanup.Push(func() (rerr error) {
					defer func() {
						cmd.Process.Kill() // no-op if already dead
						/*
							EINVAL can be returned if
							1) mountDir isn't a mountpoint (it was already unmounted)
							2) umount was called with invalid flags
							So, assuming we don't call it with invalid flags here, we are safe to ignore it.
						*/
						if err := syscall.Unmount(mountDir, syscall.MNT_FORCE); err != nil && !errors.Is(err, syscall.EINVAL) {
							rerr = multierror.Append(rerr,
								fmt.Errorf("failed to unmount fuse overlay: %w", err)).ErrorOrNil()
						}
					}()

					timeoutCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
					defer done()

					if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
						return err
					}
					select {
					case waitErr := <-waitCh:
						var exitErr *exec.ExitError
						if errors.As(waitErr, &exitErr) {
							log.Error(exitErr)
							return nil
						}
						return waitErr
					case <-timeoutCtx.Done():
						return timeoutCtx.Err()
					}
				})
				return nil
			}); err != nil {
				return nil, nil, cleanup, err
			}

			ocimounts = append(ocimounts, oci.Mount{
				Source:      mountDir,
				Destination: curTree.mountpoint,
				Type:        "none",
				Options:     []string{"bind"},
			})
			diffDirs[curTree.mountpoint] = upperDir
		}
		curTrees = nextTrees
	}
	return ocimounts, diffDirs, cleanup, nil
}

func setupPrivateDir(curTree, nextTree *tree, privateDir string) error {
	relPath, err := filepath.Rel(curTree.mountpoint, nextTree.mountpoint)
	if err != nil {
		return fmt.Errorf("failed to get rel path for private lower dir: %w", err)
	}

	privateDest := filepath.Join(privateDir, relPath)
	if nextTree.ociMount != nil && (HasBind(nextTree.ociMount.Options) || HasRBind(nextTree.ociMount.Options)) {
		if stat, err := os.Stat(nextTree.ociMount.Source); err != nil {
			return fmt.Errorf("failed to stat bind mount dest for private lower dir: %w", err)
		} else if !stat.IsDir() {
			// TODO just assuming it's a file, should handle other cases?
			parentDir := filepath.Dir(privateDest)
			err := os.MkdirAll(parentDir, 0700) // TODO set same perms
			if err != nil {
				return fmt.Errorf("failed to mkdir parent dir for private lower dir: %w", err)
			}
			err = ioutil.WriteFile(privateDest, nil, 0700) // TODO fix perms
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to touch bind mount dest for private lower dir: %w", err)
			}
			return nil
		}
	}
	if err := os.MkdirAll(privateDest, 0700); err != nil { // TODO fix perms
		return fmt.Errorf("failed to mkdir private lower dir: %w", err)
	}
	return nil
}

func waitForMountChangeAt(ctx context.Context, path string, cb func(context.Context) error) error {
	initialInfo, err := mountInfosAt(path)
	if err != nil {
		return err
	}

	if err := cb(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			newInfo, err := mountInfosAt(path)
			if err != nil {
				return err
			}
			if len(initialInfo) != len(newInfo) {
				return nil
			}
		}
	}
}

func mountInfosAt(path string) ([]mount.Info, error) {
	path = filepath.Clean(path)
	allMounts, err := mount.Self()
	if err != nil {
		return nil, err
	}
	var matching []mount.Info
	for _, mnt := range allMounts {
		if filepath.Clean(mnt.Mountpoint) == path {
			matching = append(matching, mnt)
		}
	}
	return matching, nil
}
