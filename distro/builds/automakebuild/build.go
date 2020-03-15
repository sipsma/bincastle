package automakebuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/autoconf"
	"github.com/sipsma/bincastle/distro/pkgs/automake"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
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
	automake.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	m4.Pkger
	libtool.Pkger
	perl5.Pkger
	autoconf.Pkger
}, opts ...Opt) automake.Pkg {
	return automake.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.M4(),
				d.Libtool(),
				d.Perl5(),
				d.Autoconf(),
				d.AutomakeSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/automake-src/configure`,
					`--prefix=/usr`,
					`--docdir=/usr/share/doc/automake-1.16.1`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("automake"),
			RuntimeDeps(
				d.Libc(),
				d.M4(),
				d.Perl5(),
				d.Autoconf(),
			),
		).With(opts...)
	})
}
