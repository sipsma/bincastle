package util

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
)

type LowerDir struct {
	Index     int
	Dest      string
	DiscardChanges bool
}

func (ld LowerDir) String() string {
	mountRoot := "/depends"
	if ld.DiscardChanges {
		mountRoot = "/private"
	}

	if ld.Dest == "" {
		ld.Dest = "/"
	}

	return filepath.Join(mountRoot, strconv.Itoa(ld.Index), ld.Dest)
}

func LowerDirFrom(str string) (LowerDir, error) {
	ld := LowerDir{}
	split := strings.SplitN(str, "/", 4)

	switch split[1] {
	case "depends":
	case "private":
		ld.DiscardChanges = true
	default:
		return ld, errors.New("TODO")
	}

	index, err := strconv.Atoi(split[2])
	if err != nil {
		panic("TODO")
	}
	ld.Index = index

	ld.Dest = "/"
	if len(split) == 4 {
		ld.Dest = filepath.Join(ld.Dest, split[3])
	}

	return ld, nil
}
