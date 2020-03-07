package python3build

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/bzip2"
	"github.com/sipsma/bincastle/distro/pkgs/expat"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gdbm"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libffi"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/openssl"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/python3"
	"github.com/sipsma/bincastle/distro/pkgs/readline"
	"github.com/sipsma/bincastle/distro/pkgs/xz"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	python3.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	zlib.Pkger
	bzip2.Pkger
	libffi.Pkger
	ncurses.Pkger
	gdbm.Pkger
	expat.Pkger
	openssl.Pkger
	xz.Pkger
	readline.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		zlib.Pkg(d),
		bzip2.Pkg(d),
		libffi.Pkg(d),
		ncurses.Pkg(d),
		gdbm.Pkg(d),
		expat.Pkg(d),
		openssl.Pkg(d),
		xz.Pkg(d),
		readline.Pkg(d),
		python3.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/python3-src/configure`,
				`--prefix=/usr`,
				`--enable-shared`,
				`--with-system-expat`,
				`--with-system-ffi`,
				`--with-ensurepip=yes`,
			}, " "),
			`make`,
			`make install`,
			`chmod -v 755 /usr/lib/libpython3.7m.so`,
			`chmod -v 755 /usr/lib/libpython3.so`,
			`ln -sfv pip3.7 /usr/bin/pip3`,
		),
	).With(
		Name("python3"),
		Deps(
			libc.Pkg(d),
			zlib.Pkg(d),
			bzip2.Pkg(d),
			libffi.Pkg(d),
			ncurses.Pkg(d),
			gdbm.Pkg(d),
			expat.Pkg(d),
			openssl.Pkg(d),
			xz.Pkg(d),
			readline.Pkg(d),
		),
	).With(opts...))
}
