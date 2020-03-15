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
}, opts ...Opt) perl5.Pkg {
	return perl5.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.GDBM(),
				d.Bzip2(),
				d.Zlib(),
				d.Perl5Src().With(DiscardChanges()),
			),
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
			RuntimeDeps(
				d.Libc(),
				d.GDBM(),
				d.Bzip2(),
				d.Zlib(),
			),
		).With(opts...)
	})
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
}, opts ...Opt) perl5.XMLParserPkg {
	return perl5.BuildXMLParserPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Perl5(),
				d.Expat(),
				d.Perl5XMLParserSrc(),
			),
			Shell(
				`cd /src/perl5-xmlparser-src`,
				`perl /src/perl5-xmlparser-src/Makefile.PL`,
				`make`,
				`make install`,
			),
		).With(
			Name("perl5-xmlparser"),
			RuntimeDeps(d.Libc(), d.Perl5(), d.Expat()),
		).With(opts...)
	})
}
