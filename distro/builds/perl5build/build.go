package perl5build

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/bzip2"
	"github.com/sipsma/bincastle/distro/pkgs/expat"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gdbm"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/perl5"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	perl5.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	gdbm.Pkger
	bzip2.Pkger
	zlib.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		gdbm.Pkg(d),
		bzip2.Pkg(d),
		zlib.Pkg(d),
		perl5.SrcPkg(d).With(DiscardChanges()),
		Shell(
			`cd /src/perl5-src`,
			`export BUILD_ZLIB=False`,
			`export BUILD_BZIP2=0`,
			strings.Join([]string{`sh`,
				`/src/perl5-src/Configure`,
				`-des`,
				`-Dprefix=/usr`,
				`-Dvendorprefix=/usr`,
				`-Dman1dir=/usr/share/man/man1`,
				`-Dman3dir=/usr/share/man/man3`,
				`-Dpager="/usr/bin/less -isR"`,
				`-Duseshrplib`,
				`-Dusethreads`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("perl5"),
		Deps(
			libc.Pkg(d),
			gdbm.Pkg(d),
			bzip2.Pkg(d),
			zlib.Pkg(d),
		),
	).With(opts...))
}

func DefaultXMLParser(d interface {
	PkgCache
	Executor
	perl5.XMLParserSrcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	perl5.Pkger
	expat.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		perl5.Pkg(d),
		expat.Pkg(d),
		perl5.XMLParserSrcPkg(d),
		Shell(
			`cd /src/perl5-xmlparser-src`,
			`perl /src/perl5-xmlparser-src/Makefile.PL`,
			`make`,
			`make install`,
		),
	).With(
		Name("perl5-xmlparser"),
		Deps(libc.Pkg(d), perl5.Pkg(d), expat.Pkg(d)),
	).With(opts...))
}
