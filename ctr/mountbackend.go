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
	"golang.org/x/sys/unix"
)

type MountBackend interface {
	// TODO make MountTree fields accessible outside this pkg so external impls
	// of MountBackend can be written
	SetupTree(mt *MountTree, cs ContainerState, upperdir, workdir string) ([]oci.Mount, CleanupStack, error)
}

// NoOverlayfsBackend throws an error if an overlay is required to turn the
// mount tree into a list of OCI mounts.
type NoOverlayfsBackend struct {}

func (b NoOverlayfsBackend) SetupTree(mt *MountTree, cs ContainerState, upperdir, workdir string) (_ []oci.Mount, cleanup CleanupStack, rerr error) {
	defer func() {
		if rerr != nil && len(cleanup) > 0 {
			rerr = multierror.Append(rerr, cleanup.Cleanup()).ErrorOrNil()
		}
	}()

	var ocimounts []oci.Mount

	curTrees := []*tree{mt.tree}
	for len(curTrees) > 0 {
		var nextTrees []*tree
		for _, curTree := range curTrees {
			nextTrees = append(nextTrees, curTree.submounts...)
			if len(curTree.srcs) > 0 {
				return nil, cleanup, fmt.Errorf("invalid overlayed mount: %+v", curTree.srcs)
			}
			if curTree.ociMount != nil {
				ocimounts = append(ocimounts, oci.Mount(*curTree.ociMount))
			}
		}
		curTrees = nextTrees
	}
	return ocimounts, cleanup, nil
}

type FuseOverlayfsBackend struct {
	FuseOverlayfsBin string
}

func (b FuseOverlayfsBackend) SetupTree(mt *MountTree, cs ContainerState, upperdir, workdir string) (_ []oci.Mount, cleanup CleanupStack, rerr error) {
	defer func() {
		if rerr != nil && len(cleanup) > 0 {
			rerr = multierror.Append(rerr, cleanup.Cleanup()).ErrorOrNil()
		}
	}()

	overlayDir := cs.OverlayDir()
	mountDir := overlayDir.MountDir()
	if err := os.MkdirAll(mountDir, 0700); err != nil {
		return nil, cleanup, fmt.Errorf("failed to create fuse overlay mount dir %q: %w", mountDir, err)
	}

	extraFuseOpts := []string{"clone_fd", "auto_unmount"}

	var ocimounts []oci.Mount

	curTrees := []*tree{mt.tree}
	for len(curTrees) > 0 {
		var nextTrees []*tree
		for _, curTree := range curTrees {
			nextTrees = append(nextTrees, curTree.submounts...)

			subOverlayDir := cs.SubOverlayDir(curTree.mountpoint)
			privateDir := subOverlayDir.PrivateDir()

			lowerDirs := make([]string, len(curTree.srcs))
			for i, isrc := range curTree.srcs {
				// overlay has lowerdirs from top->bottom, but we store them by
				// appending top layers to the end (bottom->top), so reverse srcs
				lowerDirs[len(curTree.srcs)-1-i] = isrc.src
			}

			if curTree.ociMount != nil {
				ociMount := oci.Mount(*curTree.ociMount)

				// If the source of a bind-mount is not already read-only but the bind-mount is
				// read-only (or vice versa), the mount can fail under some circumstances w/
				// EPERM. This section checks for that situation and works around it by not
				// directly bind mounting but instead creating a read-only fuse mount and then
				// bind-mounting that fuse-mount instead.
				if HasAnyBind(curTree.ociMount.Options) {
					var statfs unix.Statfs_t
					if err := unix.Statfs(curTree.ociMount.Source, &statfs); err != nil {
						return nil, cleanup, fmt.Errorf("failed to statfs bind mount source %q: %w",
							curTree.ociMount.Source, err)
					}

					if HasReadOnly(curTree.ociMount.Options) && (statfs.Flags&unix.ST_RDONLY == 0) {
						subMountDir := subOverlayDir.MountDir()
						if err := os.MkdirAll(subMountDir, 0700); err != nil {
							return nil, cleanup, fmt.Errorf(
								"failed to create submount dir %q: %w", subMountDir, err)
						}

						ociMount = oci.Mount{
							Source:      subMountDir,
							Destination: curTree.mountpoint,
							Type:        "none",
							Options:     ReplaceOption(curTree.ociMount.Options, "ro", ""),
						}

						stat, err := os.Stat(curTree.ociMount.Source)
						if err != nil {
							return nil, cleanup, err
						}
						if stat.IsDir() {
							lowerDirs = []string{curTree.ociMount.Source}
						} else {
							parentDir := filepath.Dir(curTree.ociMount.Source)
							lowerDirs = []string{parentDir}
							ociMount.Source = filepath.Join(ociMount.Source, filepath.Base(curTree.ociMount.Source))
						}

						if cleanupCmd, err := b.runCmd(
							lowerDirs, "", "", append(extraFuseOpts, "ro"), subMountDir,
						); err != nil {
							return nil, cleanup, err
						} else {
							cleanup = cleanup.Push(cleanupCmd)
						}
					}
				}
				ocimounts = append(ocimounts, ociMount)
				continue
			}
			if len(lowerDirs) == 0 {
				continue
			}

			if err := os.MkdirAll(privateDir, 0700); err != nil {
				return nil, cleanup, fmt.Errorf(
					"failed to create private lower dir: %w", err)
			}
			cleanup = append(cleanup, func() error {
				return os.RemoveAll(privateDir)
			})

			// setup private lower dir, which ensures that mounts under this one have
			// a mountpoint w/out needing to change any lowerdirs here
			for _, nextTree := range nextTrees {
				if err := setupPrivateDir(curTree, nextTree, privateDir); err != nil {
					return nil, cleanup, err
				}
			}

			lowerDirs = append(lowerDirs, privateDir)

			subMountDir := filepath.Join(mountDir, curTree.mountpoint)
			if err := os.MkdirAll(subMountDir, 0700); err != nil {
				return nil, cleanup, fmt.Errorf(
					"failed to create submount dir %q: %w", subMountDir, err)
			}

			if cleanupCmd, err := b.runCmd(lowerDirs, "", "", append(extraFuseOpts, "ro"), subMountDir); err != nil {
				return nil, cleanup, err
			} else {
				cleanup = cleanup.Push(cleanupCmd)
			}
		}
		curTrees = nextTrees
	}

	if upperdir != "" {
		if err := os.MkdirAll(upperdir, 0700); err != nil {
			return nil, cleanup, fmt.Errorf(
				"failed to create upper dir %q: %w", upperdir, err)
		}
		if err := os.MkdirAll(workdir, 0700); err != nil {
			return nil, cleanup, fmt.Errorf(
				"failed to create work dir %q: %w", workdir, err)
		}

		if cleanupCmd, err := b.runCmd([]string{mountDir}, upperdir, workdir, extraFuseOpts, cs.rootfsDir()); err != nil {
			return nil, cleanup, err
		} else {
			cleanup = cleanup.Push(cleanupCmd)
		}

		cleanup = cleanup.Push(func() error {
			// If a directory is created in the upperdir without a corresponding path in a lowerdir, then
			// it will be set as opaque. This is not desired once the upperdir is finished and is ready
			// to become a lowerdir for a future mount, so de-opaque all directories in the upperdir.
			return filepath.Walk(upperdir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					// safe to ignore error; it either means there already wasn't such an xattr
					// or the fs doesn't support xattrs in the first place
					_ = unix.Removexattr(path, "user.fuseoverlayfs.opaque")
					return nil
				}
				if filepath.Base(path) == ".wh..wh..opq" {
					return os.Remove(path)
				}
				return nil
			})
		})
	} else {
		ocimounts = append([]oci.Mount{{
			Source:      mountDir,
			Destination: "/",
			Type:        "none",
			Options:     []string{"rbind"},
		}}, ocimounts...)
	}
	return ocimounts, cleanup, nil
}

func (b FuseOverlayfsBackend) runCmd(
	lowerdirs []string,
	upperdir string,
	workdir string,
	extraOpts []string,
	mountdir string,
) (func() error, error) {
	cmd := exec.Cmd{
		Path: b.FuseOverlayfsBin,
		Args: []string{b.FuseOverlayfsBin,
			"-f", // stay in foreground
			"-o", overlayOptions{
				LowerDirs: lowerdirs,
				UpperDir:  upperdir,
				WorkDir:   workdir,
				Extra:     extraOpts,
			}.Options(),
			mountdir,
		},
		// TODO optional way to enable debug mode and send output somewhere?
		// Stdout: os.Stdout,
		// Stderr: os.Stderr,
	}

	timeoutCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
	defer done()

	var cleanup func() error
	if err := waitForMountChangeAt(timeoutCtx, mountdir, func(context.Context) error {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start fuse-overlayfs for mount dir %s: %w", mountdir, err)
		}

		waitCh := make(chan error)
		go func() {
			defer close(waitCh)
			waitCh <- cmd.Wait()
		}()

		cleanup = func() (rerr error) {
			defer cmd.Process.Kill() // no-op if already dead

			timeoutCtx, done := context.WithTimeout(context.Background(), 10*time.Second)
			defer done()

			return waitForMountChangeAt(timeoutCtx, mountdir, func(ctx context.Context) error {
				if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
					return err
				}
				select {
				case waitErr := <-waitCh:
					var exitErr *exec.ExitError
					if errors.As(waitErr, &exitErr) {
						log.Error(exitErr)
						return nil // TODO distinguish error of receiving SIGTERM vs unexpected?
					}
					return waitErr
				case <-ctx.Done():
					return ctx.Err()
				}
			})
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to create fuse overlay at %q: %w", mountdir, err)
	}
	return cleanup, nil
}

func setupPrivateDir(curTree, nextTree *tree, privateDir string) error {
	relPath, err := filepath.Rel(curTree.mountpoint, nextTree.mountpoint)
	if err != nil {
		return fmt.Errorf("failed to get rel path for private lower dir: %w", err)
	}

	privateDest := filepath.Join(privateDir, relPath)
	if nextTree.ociMount != nil && HasAnyBind(nextTree.ociMount.Options) {
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

	ticker := time.NewTicker(50 * time.Millisecond)
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
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}

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
