package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Perl5 struct{}

func (Perl5) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Bzip2{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(M4{}),
		BuildDep(src.Perl5{}),
		ScratchMount(`/build`),
		bootstrap.BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{`sh`, `/src/perl5-src/Configure`,
				`-des`,
				`-Dmksymlinks`,
				`-Dprefix=/tools`,
				`-Dlibs=-lm`,
				`-Uloclibpth`,
				`-Ulocincpth`,
			}, " "),
			`make`,
			`cp -Lv perl cpan/podlators/scripts/pod2man /tools/bin`,
			`mkdir -pv /tools/lib/perl5/5.30.0`,
			`cp -RLv lib/* /tools/lib/perl5/5.30.0`,
		),
	)
}
