package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type UtilLinux struct{}

func (UtilLinux) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		Dep(Readline{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.UtilLinux{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			`mkdir -pv /var/lib/hwclock`,
			strings.Join([]string{
				`/src/utillinux-src/configure`,
				`ADJTIME_PATH=/var/lib/hwclock/adjtime`,
				`--docdir=/usr/share/doc/util-linux-2.34`,
				`--disable-chfn-chsh`,
				`--disable-login`,
				`--disable-nologin`,
				`--disable-su`,
				`--disable-setpriv`,
				`--disable-runuser`,
				`--disable-pylibmount`,
				`--disable-static`,
				`--without-python`,
				`--without-systemd`,
				`--without-systemdsystemunitdir`,
				// TODO below are only to avoid errors in userns
				`--disable-use-tty-group`,
				`--disable-makeinstall-chown`,
				`--disable-makeinstall-setuid`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
