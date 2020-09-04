package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Perl5 struct{}

func (Perl5) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(GDBM{}),
		Dep(Bzip2{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Perl5{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			`export BUILD_ZLIB=False`,
			`export BUILD_BZIP2=0`,
			strings.Join([]string{`sh`,
				`/src/perl5-src/Configure`,
				`-des`,
				`-Dmksymlinks`,
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
	)
}

type Perl5XMLParser struct{}

func (Perl5XMLParser) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Perl5{}),
		Dep(Expat{}),
		BuildDep(LayerSpec(
			Dep(src.Perl5XMLParser{}),
			BuildDep(Libc{}),
			BuildDep(Perl5{}),
			BuildDep(Expat{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(Binutils{}),
			BuildDep(GCC{}),
			BuildDep(PkgConfig{}),
			BuildOpts(),
			BuildScript(
				`cd /src/perl5-xmlparser-src`,
				`perl /src/perl5-xmlparser-src/Makefile.PL`,
				`make`,
			),
		)),
		BuildOpts(),
		BuildScript(
			`cd /src/perl5-xmlparser-src`,
			`make install`,
		),
	)
}
