package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type tmpBinutils struct{}

func (tmpBinutils) Spec() Spec {
	return LayerSpec(
		Dep(bootstrap.Spec{}),
		BuildDep(src.Binutils{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/binutils-src/configure`,
				`--prefix=/tools`,
				`--with-sysroot=/sysroot`,
				`--with-lib-path=/tools/lib`,
				`--target=x86_64-bincastle-linux-gnu`,
				`--disable-nls`,
				`--disable-werror`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}

type tmpGCC struct{}

func (tmpGCC) Spec() Spec {
	return LayerSpec(
		Dep(bootstrap.Spec{}),
		Dep(tmpBinutils{}),
		BuildDep(LayerSpec(
			Dep(src.GCC{}),
			BuildDep(bootstrap.Spec{}),
			bootstrap.BuildOpts(),
			BuildScript(
				`cd /src/gcc-src`,
				`for file in gcc/config/{linux,i386/linux{,64}}.h`,
				`do`,
				`cp -uv $file{,.orig}`,
				`sed -e 's@/lib\(64\)\?\(32\)\?/ld@/tools&@g' -e 's@/usr@/tools@g' $file.orig > $file`,
				`echo '' >> $file`,
				`echo '#undef STANDARD_STARTFILE_PREFIX_1' >> $file`,
				`echo '#undef STANDARD_STARTFILE_PREFIX_2' >> $file`,
				`echo '#define STANDARD_STARTFILE_PREFIX_1 "/tools/lib/"' >> $file`,
				`echo '#define STANDARD_STARTFILE_PREFIX_2 ""' >> $file`,
				`touch $file.orig`,
				`done`,
				`sed -e '/m64=/s/lib64/lib/' -i.orig gcc/config/i386/t-linux64`,
				// TODO use Mountdir instead of linking here
				`ln -s /src/mpfr-src /src/gcc-src/mpfr`,
				`ln -s /src/gmp-src /src/gcc-src/gmp`,
				`ln -s /src/mpc-src /src/gcc-src/mpc`,
			),
		)),
		BuildDep(src.MPFR{}),
		BuildDep(src.GMP{}),
		BuildDep(src.MPC{}),
		bootstrap.BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/gcc-src/configure`,
				`--target=x86_64-bincastle-linux-gnu`,
				`--prefix=/tools`,
				`--with-glibc-version=2.11`,
				`--with-sysroot=/sysroot`,
				`--with-newlib`,
				`--without-headers`,
				`--with-local-prefix=/tools`,
				`--with-native-system-header-dir=/tools/include`,
				`--disable-nls`,
				`--disable-shared`,
				`--disable-multilib`,
				`--disable-decimal-float`,
				`--disable-threads`,
				`--disable-libatomic`,
				`--disable-libgomp`,
				`--disable-libquadmath`,
				`--disable-libssp`,
				`--disable-libvtv`,
				`--disable-libstdcxx`,
				`--enable-languages=c,c++`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}

type Libc struct{}

func (Libc) Spec() Spec {
	return LayerSpec(
		Dep(bootstrap.Spec{}),
		Dep(tmpBinutils{}),
		Dep(tmpGCC{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(src.Libc{}),
		BuildScratch(`/build`),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/libc-src/configure`,
				`--prefix=/tools`,
				`--host=x86_64-bincastle-linux-gnu`,
				`--build=$(/src/libc-src/scripts/config.guess)`,
				`--enable-kernel=3.2`,
				`--with-headers=/tools/include`,
			}, ` `),
			`make`,
			`make install`,
		),
	)
}
