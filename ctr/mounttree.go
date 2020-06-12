package ctr

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	oci "github.com/opencontainers/runtime-spec/specs-go"
)

type Mounts []MountTreeOpt

func (mounts Mounts) With(more ...MountTreeOpt) Mounts {
	return append(mounts, more...)
}

func (mounts Mounts) OCIMounts(state ContainerState) ([]oci.Mount, error) {
	mtree := &MountTree{tree: &tree{
		mountpoint: "/",
	}}
	for _, opt := range mounts {
		if err := opt.AddToMountTree(mtree); err != nil {
			return nil, err
		}
	}
	return mtree.OCIMounts(state), nil
}

type MountTree struct {
	*tree
	index int
}

func (t *MountTree) OCIMounts(state ContainerState) []oci.Mount {
	var ocimounts []oci.Mount
	curTrees := []*tree{t.tree}
	for len(curTrees) > 0 {
		var nextTrees []*tree
		for _, curTree := range curTrees {
			if ocimount := curTree.toOCIMount(state); ocimount != nil {
				ocimounts = append(ocimounts, *ocimount)
			}
			nextTrees = append(nextTrees, curTree.submounts...)
		}
		curTrees = nextTrees
	}
	return ocimounts
}

type tree struct {
	mountpoint string
	submounts  []*tree
	srcs       []indexedSrc
	ociMount   *OCIMount
}

func (t *tree) toOCIMount(state ContainerState) *oci.Mount {
	if t.ociMount != nil {
		return (*oci.Mount)(t.ociMount)
	}

	if len(t.srcs) == 0 {
		return nil
	}

	// overlay has lowerdirs from top->bottom, but we store them by
	// appending top layers to the end (bottom->top), so reverse srcs
	lowerdirs := make([]string, len(t.srcs))
	for i, isrc := range t.srcs {
		lowerdirs[len(t.srcs)-1-i] = isrc.src
	}
	overlayDir := state.OverlayDir(t.mountpoint)
	return &oci.Mount{
		Source:      "none",
		Destination: t.mountpoint,
		Type:        "overlay",
		Options: overlayOptions{
			LowerDirs: lowerdirs,
			UpperDir:  overlayDir.UpperDir(),
			WorkDir:   overlayDir.WorkDir(),
		}.OptionsSlice(),
	}
}

type indexedSrc struct {
	src   string
	index int
}

type MountTreeOpt interface {
	AddToMountTree(*MountTree) error
}

type Layer struct {
	Src  string
	Dest string
	// TODO support overlay mount opts
}

type InvalidMergedMountErr struct{}

func (e InvalidMergedMountErr) Error() string {
	// TODO
	return ""
}

func (l Layer) AddToMountTree(t *MountTree) error {
	src, err := filepath.EvalSymlinks(l.Src)
	if err != nil {
		return err
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return err
	}
	isrc := indexedSrc{src: src, index: t.index}
	t.index++

	// TODO cleanup l.Dest too

	// TODO verify that l.Src isn't a file? files should use an OCIMount (which can't be merged)

	curTree := t.tree
	for {
		if l.Dest == curTree.mountpoint {
			if curTree.ociMount != nil {
				return InvalidMergedMountErr{}
			}
			curTree.srcs = append(curTree.srcs, isrc)
			return propogateSrc(curTree, isrc)
		}

		var nextTree *tree
		for _, subtree := range curTree.submounts {
			if l.Dest == subtree.mountpoint || isUnderDir(l.Dest, subtree.mountpoint) {
				nextTree = subtree
				break
			}
		}
		if nextTree != nil {
			curTree = nextTree
			continue
		}

		parentTree := curTree
		curTree = &tree{
			mountpoint: l.Dest,
			srcs:       []indexedSrc{isrc},
		}

		rel, err := filepath.Rel(parentTree.mountpoint, curTree.mountpoint)
		if err != nil {
			return err
		}
		for _, parentSrc := range parentTree.srcs {
			overlap := filepath.Join(parentSrc.src, rel)
			if _, err := os.Lstat(overlap); err == nil {
				curTree.srcs = append(curTree.srcs, indexedSrc{
					src:   overlap,
					index: parentSrc.index,
				})
			}
		}
		sort.Slice(curTree.srcs, func(i, j int) bool {
			return curTree.srcs[i].index < curTree.srcs[j].index
		})

		newParentSubmounts := []*tree{curTree}
		for _, parentSubmount := range parentTree.submounts {
			if isUnderDir(parentSubmount.mountpoint, curTree.mountpoint) {
				curTree.submounts = append(curTree.submounts, parentSubmount)
			} else {
				newParentSubmounts = append(newParentSubmounts, parentSubmount)
			}
		}
		parentTree.submounts = newParentSubmounts

		return propogateSrc(curTree, isrc)
	}
}

func propogateSrc(t *tree, isrc indexedSrc) error {
	for _, subtree := range t.submounts {
		rel, err := filepath.Rel(t.mountpoint, subtree.mountpoint)
		if err != nil {
			return err
		}
		overlap := filepath.Join(isrc.src, rel)
		if _, err := os.Lstat(overlap); err != nil {
			continue
		}
		subISrc := indexedSrc{src: overlap, index: isrc.index}
		subtree.srcs = append(subtree.srcs, subISrc)
		propogateSrc(subtree, subISrc)
	}
	return nil
}

type OCIMount oci.Mount

func (m OCIMount) AddToMountTree(t *MountTree) error {
	// TODO cleanup m.Destination
	curTree := t.tree
	for {
		if m.Destination == curTree.mountpoint {
			return InvalidMergedMountErr{}
		}

		var nextTree *tree
		for _, subtree := range curTree.submounts {
			if m.Destination == subtree.mountpoint || isUnderDir(m.Destination, subtree.mountpoint) {
				nextTree = subtree
				break
			}
		}
		if nextTree != nil {
			curTree = nextTree
			continue
		}

		parentTree := curTree
		curTree = &tree{
			mountpoint: m.Destination,
			ociMount:   &m,
		}

		newParentSubmounts := []*tree{curTree}
		for _, parentSubmount := range parentTree.submounts {
			if isUnderDir(parentSubmount.mountpoint, curTree.mountpoint) {
				curTree.submounts = append(curTree.submounts, parentSubmount)
			} else {
				newParentSubmounts = append(newParentSubmounts, parentSubmount)
			}
		}
		parentTree.submounts = newParentSubmounts

		return nil
	}
}

func isUnderDir(path string, baseDir string) bool {
	path = filepath.Clean(path)
	baseDir = filepath.Clean(baseDir)
	if baseDir == "/" && path != "/" {
		return true
	}
	return strings.HasPrefix(path, baseDir+"/")
}
