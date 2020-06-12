package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type GCC struct{}

func (GCC) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Binutils{}),
		Dep(MPC{}),
		Dep(GMP{}),
		Dep(MPFR{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(LayerSpec(
			Dep(src.GCC{}),
			BuildDep(bootstrap.Spec{}),
			BuildOpts(),
			Shell(
				`cd /src/gcc-src`,
				`sed -e '/m64=/s/lib64/lib/' -i.orig gcc/config/i386/t-linux64`,
			),
		)),
		ScratchMount(`/build`),
		BuildOpts(),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`SED=sed`,
				`/src/gcc-src/configure`,
				`--prefix=/usr`,
				`--enable-languages=c,c++`,
				`--disable-multilib`,
				`--disable-bootstrap`,
				`--with-system-zlib`,
			}, " "),
			`make`,
			`make install`,
			`rm -rf /usr/lib/gcc/$(gcc -dumpmachine)/9.2.0/include-fixed/bits/`,
			// TODO don't hardcode uid/gid?
			`chown -v -R 0:0 /usr/lib/gcc/*linux-gnu/9.2.0/include{,-fixed}`,
			`ln -sv ../usr/bin/cpp /lib`,
			`ln -sv gcc /usr/bin/cc`,
			`install -v -dm755 /usr/lib/bfd-plugins`,
			`ln -sfv ../../libexec/gcc/$(gcc -dumpmachine)/9.2.0/liblto_plugin.so  /usr/lib/bfd-plugins/`,
			`mkdir -pv /usr/share/gdb/auto-load/usr/lib`,
			`mv -v /usr/lib/*gdb.py /usr/share/gdb/auto-load/usr/lib`,
		),
	)
}
