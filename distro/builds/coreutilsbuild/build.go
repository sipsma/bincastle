package coreutilsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/automake"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/coreutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libcap"
	"github.com/sipsma/bincastle/distro/pkgs/libtool"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	coreutils.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	automake.Pkger
	libtool.Pkger
	acl.Pkger
	attr.Pkger
	libcap.Pkger
	gmp.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		automake.Pkg(d),
		libtool.Pkg(d),
		acl.Pkg(d),
		attr.Pkg(d),
		libcap.Pkg(d),
		gmp.Pkg(d),
		coreutils.SrcPkg(d).With(DiscardChanges()),
		Shell(
			`cd /src/coreutils-src`,
			`autoreconf -fiv`,
			strings.Join([]string{
				`/src/coreutils-src/configure`,
				`--prefix=/usr`,
				`--enable-no-install-program=kill,uptime`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/bin/{cat,chgrp,chmod,chown,cp,date,dd,df,echo} /bin`,
			`mv -v /usr/bin/{false,ln,ls,mkdir,mknod,mv,pwd,rm} /bin`,
			`mv -v /usr/bin/{rmdir,stty,sync,true,uname} /bin`,
			`mv -v /usr/bin/chroot /usr/sbin`,
			`mv -v /usr/share/man/man1/chroot.1 /usr/share/man/man8/chroot.8`,
			`sed -i s/\"1\"/\"8\"/1 /usr/share/man/man8/chroot.8`,
			`mv -v /usr/bin/{head,nice,sleep,touch} /bin`,
		),
	).With(
		Name("coreutils"),
		VersionOf(coreutils.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			acl.Pkg(d),
			attr.Pkg(d),
			libcap.Pkg(d),
			gmp.Pkg(d),
		),
	).With(opts...))
}