package src

import (
	"fmt"
	"path/filepath"

	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

type URL string
type urlKey struct{}
func URLOf(p Pkg) string {
	url, ok := PkgValueOf(p, urlKey{}).(string)
	if !ok {
		return ""
	}
	return url
}
func (u URL) ApplyToPkg(p Pkg) Pkg {
	return PkgValue(urlKey{}, string(u)).ApplyToPkg(p)
}

const defaultStripComponents = 1
type StripComponents int
type stripComponentsKey struct{}
func StripComponentsOf(p Pkg) int {
	stripComponents, ok := PkgValueOf(p, stripComponentsKey{}).(int)
	if !ok {
		return defaultStripComponents
	}
	return stripComponents
}
func (s StripComponents) ApplyToPkg(p Pkg) Pkg {
	return PkgValue(stripComponentsKey{}, int(s)).ApplyToPkg(p)
}

func Curl(distro Executor, opts ...Opt) PkgBuild {
	opt := PkgOf(opts...)
	url := URLOf(opt)
	if url == "" {
		panic("TODO")
	}

	name := NameOf(opt)
	if name == "" {
		panic("TODO use a random default instead of panicking?")
	}

	stripComponents := StripComponentsOf(opt)

	return PkgBuildOf(distro.Exec(Shell(
		`mkdir -p /src`,
		`cd /src`,
		fmt.Sprintf("curl -L -O %s", url),
		`DLFILE=$(ls)`,
		fmt.Sprintf(
			`tar --strip-components=%d --extract --no-same-owner --file=$DLFILE`,
			stripComponents),
		`rm $DLFILE`,
	)).With(
		OutputDir("/src"),
		MountDir(filepath.Join("/src", name)),
	).With(opts...))
}

type Ref string
type refKey struct{}
const defaultRef = "master"
func RefOf(p Pkg) string {
	ref, ok := PkgValueOf(p, refKey{}).(string)
	if !ok {
		return defaultRef
	}
	return ref
}
func (r Ref) ApplyToPkg(p Pkg) Pkg {
	return PkgValue(refKey{}, string(r)).ApplyToPkg(p)
}

func Git(distro Executor, opts ...Opt) PkgBuild {
	opt := PkgOf(opts...)
	url := URLOf(opt)
	if url == "" {
		panic("TODO")
	}

	name := NameOf(opt)
	if name == "" {
		panic("TODO use a random default instead of panicking?")
	}

	ref := RefOf(opt)

	return PkgBuildOf(distro.Exec(Shell(
		`mkdir -p /src`,
		fmt.Sprintf(`git clone --recurse-submodules %s /src`, url),
		`cd /src`,
		fmt.Sprintf(`git checkout %s`, ref),
	)).With(
		OutputDir("/src"),
		MountDir(filepath.Join("/src", name)),
	).With(opts...))
}
