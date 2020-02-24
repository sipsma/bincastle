package bcbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/bc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	bc.Srcer
	libc.Pkger
	linux.HeadersPkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		bc.SrcPkg(d).With(DiscardChanges()),
		Shell(
			`cd /src/bc-src`,
			strings.Join([]string{
				`PREFIX=/usr`,
				`CC=gcc`,
				`CFLAGS="-std=c99"`,
				`./configure.sh`,
				`-G`,
				`-O3`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("bc"),
		VersionOf(bc.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	).With(opts...))
}
