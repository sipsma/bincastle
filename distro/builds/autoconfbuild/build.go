package autoconfbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/autoconf"
	"github.com/sipsma/bincastle/distro/pkgs/bash"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gettext"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libtool"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	"github.com/sipsma/bincastle/distro/pkgs/perl5"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	autoconf.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	m4.Pkger
	libtool.Pkger
	perl5.Pkger
	gettext.Pkger
	bash.Pkger
}, opts ...Opt) autoconf.Pkg {
	return autoconf.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.M4(),
				d.Libtool(),
				d.Perl5(),
				d.Gettext(),
				d.Bash(),
				Patch(d, d.AutoconfSrc(), Shell(
					`cd /src/autoconf-src`,
					`sed '361 s/{/\\{/' -i bin/autoscan.in`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/autoconf-src/configure`,
					`--prefix=/usr`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("autoconf"),
			RuntimeDeps(
				d.Libc(),
				d.M4(),
				d.Libtool(),
				d.Perl5(),
				d.Gettext(),
				d.Bash(),
			),
		).With(opts...)
	})
}
