package libcbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/timezonedata"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func DefaultGlibc(d interface {
	PkgCache
	Executor
	// TODO should ask for GlibcSrc specifically, right?
	// as opposed to musl or something like that
	libc.Srcer
	timezonedata.Srcer
	linux.HeadersPkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		timezonedata.SrcPkg(d),
		Patch(d, libc.SrcPkg(d), Shell(
			`cd /src/libc-src`,
			// TODO is this needed?
			`sed -i '/asm.socket.h/a# include <linux/sockios.h>' sysdeps/unix/sysv/linux/bits/socket.h`,
		)),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			`ln -sfv ../lib/ld-linux-x86-64.so.2 /lib64`,
			`ln -sfv ../lib/ld-linux-x86-64.so.2 /lib64/ld-lsb-x86-64.so.3`,
			strings.Join([]string{
				`CC="gcc -ffile-prefix-map=/tools=/usr"`,
				`/src/libc-src/configure`,
				`--prefix=/usr`,
				`--disable-werror`,
				`--enable-kernel=3.2`,
				`--enable-stack-protector=strong`,
				`--with-headers=/usr/include`,
				`libc_cv_slibdir=/lib`,
			}, " "),
			`make`,
			`sed '/test-installation/s@$(PERL)@echo not running@' -i /src/libc-src/Makefile`,
			`make install`,
			`cp -v /src/libc-src/nscd/nscd.conf /etc/nscd.conf`,
			`mkdir -pv /var/cache/nscd`,
			`mkdir -pv /usr/lib/locale`,
			`echo 'passwd: files' > /etc/nsswitch.conf`,
			`echo 'group: files' >> /etc/nsswitch.conf`,
			`echo 'shadow: files' >> /etc/nsswitch.conf`,
			`echo 'hosts: files dns' >> /etc/nsswitch.conf`,
			`echo 'networks: files' >> /etc/nsswitch.conf`,
			`echo 'protocols: files' >> /etc/nsswitch.conf`,
			`echo 'services: files' >> /etc/nsswitch.conf`,
			`echo 'ethers: files' >> /etc/nsswitch.conf`,
			`echo 'rpc: files' >> /etc/nsswitch.conf`,
			`localedef -i POSIX -f UTF-8 C.UTF-8 2> /dev/null || true`,
			`localedef -i en_US -f ISO-8859-1 en_US`,
			`localedef -i en_US -f UTF-8 en_US.UTF-8`,
			`mkdir -pv /usr/share/zoneinfo/{posix,right}`,
			`for tz in etcetera southamerica northamerica europe africa antarctica asia australasia backward pacificnew systemv; do`,
			`zic -L /dev/null -d /usr/share/zoneinfo /src/timezonedata-src/${tz}`,
			`zic -L /dev/null -d /usr/share/zoneinfo/posix /src/timezonedata-src/${tz}`,
			`zic -L /src/timezonedata-src/leapseconds -d /usr/share/zoneinfo/right /src/timezonedata-src/${tz}`,
			`done`,
			`cp -v /src/timezonedata-src/zone.tab /usr/share/zoneinfo`,
			`cp -v /src/timezonedata-src/zone1970.tab /usr/share/zoneinfo`,
			`cp -v /src/timezonedata-src/iso3166.tab /usr/share/zoneinfo`,
			`zic -d /usr/share/zoneinfo -p America/New_York`,
			`ln -sfv /usr/share/zoneinfo/America/Los_Angeles /etc/localtime`,
			`echo '/usr/local/lib' > /etc/ld.so.conf`,
			`echo '/opt/lib' >> /etc/ld.so.conf`,
			`echo 'include /etc/ld.so.conf.d/*.conf' >> /etc/ld.so.conf`,
			`mkdir -pv /etc/ld.so.conf.d`,
		),
	).With(
		Name("libc"),
		VersionOf(libc.SrcPkg(d)),
	).With(opts...))
}
