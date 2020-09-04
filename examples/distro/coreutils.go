package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Coreutils struct{}

func (Coreutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Acl{}),
		Dep(Attr{}),
		Dep(Libcap{}),
		Dep(GMP{}),
		BuildDep(Autoconf{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(LayerSpec(
			Dep(src.Coreutils{}),
			BuildDep(Libc{}),
			BuildDep(Acl{}),
			BuildDep(Attr{}),
			BuildDep(Libcap{}),
			BuildDep(GMP{}),
			BuildDep(LinuxHeaders{}),
			BuildDep(GCC{}),
			BuildDep(Binutils{}),
			BuildDep(PkgConfig{}),
			BuildDep(Automake{}),
			BuildDep(Autoconf{}),
			BuildDep(M4{}),
			BuildDep(Libtool{}),
			BuildOpts(),
			BuildScript(
				`cd /src/coreutils-src`,
				`autoreconf -fiv`,
				strings.Join([]string{
					`/src/coreutils-src/configure`,
					`--prefix=/usr`,
					`--enable-no-install-program=kill,uptime`,
				}, " "),
				`make`,
			),
		)),
		BuildOpts(),
		BuildScript(
			`cd /src/coreutils-src`,
			`make install`,
			`mv -v /usr/bin/{cat,chgrp,chmod,chown,cp,date,dd,df,echo} /bin`,
			`mv -v /usr/bin/{false,ln,ls,mkdir,mknod,mv,pwd,rm} /bin`,
			`mv -v /usr/bin/{rmdir,stty,sync,true,uname} /bin`,
			`mv -v /usr/bin/chroot /usr/sbin`,
			`mv -v /usr/share/man/man1/chroot.1 /usr/share/man/man8/chroot.8`,
			`sed -i s/\"1\"/\"8\"/1 /usr/share/man/man8/chroot.8`,
			`mv -v /usr/bin/{head,nice,sleep,touch} /bin`,
		),
	)
}
