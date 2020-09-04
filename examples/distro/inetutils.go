package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Inetutils struct{}

func (Inetutils) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Readline{}),
		Dep(Ncurses{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Inetutils{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/inetutils-src/configure`,
				`--prefix=/usr`,
				`--localstatedir=/var`,
				`--disable-logger`,
				`--disable-whois`,
				`--disable-rcp`,
				`--disable-rexec`,
				`--disable-rlogin`,
				`--disable-rsh`,
				`--disable-servers`,
			}, " "),
			`make`,
			`make install`,
			`mv -v /usr/bin/{hostname,ping,ping6,traceroute} /bin`,
			`mv -v /usr/bin/ifconfig /sbin`,
		),
	)
}
