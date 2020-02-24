package golangbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/golang"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

// TODO how to version builds better
func Golang(d interface {
	PkgCache
	Executor
	golang.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		golang.SrcPkg(d).With(SwapToVersion(V(1,4))),
		golang.SrcPkg(d),
		Shell(
			`cd /src/golang1.4-src/src`,
			`./make.bash`,
			`cd /src/golang1.13-src/src`,
			strings.Join([]string{
				`GOROOT_BOOTSTRAP=/src/golang1.4-src`,
				`GOROOT_FINAL=/usr/lib/go`,
				`GOBIN=/usr/bin`,
				`./make.bash`,
			}, " "),
			`mkdir -p /usr/lib/go`,
			`mv /src/golang1.13-src/* /usr/lib/go`,
		),
	).With(
		Name("golang"),
		VersionOf(golang.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
