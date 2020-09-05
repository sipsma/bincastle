package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type libstdcpp struct{}

func (libstdcpp) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(src.GCC{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildScratch(`/build`),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/gcc-src/libstdc++-v3/configure`,
				`--host=x86_64-bincastle-linux-gnu`,
				`--prefix=/tools`,
				`--disable-multilib`,
				`--disable-nls`,
				`--disable-libstdcxx-threads`,
				`--disable-libstdcxx-pch`,
				`--with-gxx-include-dir=/tools/x86_64-bincastle-linux-gnu/include/c++/9.2.0`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}

type GCC struct{}

func (GCC) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(libstdcpp{}),
		Dep(Binutils{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
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
			// TODO don't do this here
			`cd /src/gcc-src`,
			"cat gcc/limitx.h gcc/glimits.h gcc/limity.h > `dirname $(x86_64-bincastle-linux-gnu-gcc -print-libgcc-file-name)`/include-fixed/limits.h",

			`cd /build`,
			strings.Join([]string{
				`CC=x86_64-bincastle-linux-gnu-gcc`,
				`CXX=x86_64-bincastle-linux-gnu-g++`,
				`AR=x86_64-bincastle-linux-gnu-ar`,
				`RANLIB=x86_64-bincastle-linux-gnu-ranlib`,
				`/src/gcc-src/configure`,
				`--prefix=/tools`,
				`--with-local-prefix=/tools`,
				`--with-native-system-header-dir=/tools/include`,
				`--enable-languages=c,c++`,
				`--disable-libstdcxx-pch`,
				`--disable-multilib`,
				`--disable-bootstrap`,
				`--disable-libgomp`,
			}, " "),
			`make`,
			`make install`,
			`ln -sv gcc /tools/bin/cc`,
		),
	)
}
