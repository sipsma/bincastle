package util

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/moby/buildkit/client/llb"
)

// TODO don't hardcode bash
func Shell(lines ...string) llb.RunOption {
	lines = append([]string{`set -e`, `set +h`}, lines...)
	return llb.Args([]string{
		"bash", "-e", "-c", fmt.Sprintf(
			// TODO using THEREALEOF allows callers to use <<EOF in their
			// own shell lines, but is there a better way?
			"exec bash <<\"THEREALEOF\"\n%s\nTHEREALEOF",
			strings.Join(lines, "\n"),
		),
	})
}

func ScratchMount(dest string) llb.RunOption {
	return llb.AddMount(dest, llb.Scratch(), llb.ForceNoOutput)
}

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

	if len(split) == 4 {
		ld.Dest = split[3]
	} else {
		ld.Dest = "/"
	}

	return ld, nil
}
