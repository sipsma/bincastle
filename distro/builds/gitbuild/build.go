package gitbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/cacerts"
	"github.com/sipsma/bincastle/distro/pkgs/curl"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/git"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	git.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	m4.Pkger
	zlib.Pkger
	curl.Pkger
	cacerts.Pkger
	// TODO libpcre?
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		zlib.Pkg(d),
		curl.Pkg(d),
		git.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /src/git-src`,
			strings.Join([]string{
				`/src/git-src/configure`,
				`--prefix=/usr`,
				`--with-gitconfig=/etc/gitconfig`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("git"),
		VersionOf(git.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			zlib.Pkg(d),
			cacerts.Pkg(d),
			curl.Pkg(d),
		),
	).With(opts...))
}
