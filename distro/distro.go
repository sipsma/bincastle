package distro

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/moby/buildkit/client/llb"

	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"

	"github.com/sipsma/bincastle/distro/builds/aclbuild"
	"github.com/sipsma/bincastle/distro/builds/emacsbuild"
	"github.com/sipsma/bincastle/distro/builds/attrbuild"
	"github.com/sipsma/bincastle/distro/builds/autoconfbuild"
	"github.com/sipsma/bincastle/distro/builds/automakebuild"
	"github.com/sipsma/bincastle/distro/builds/awkbuild"
	"github.com/sipsma/bincastle/distro/builds/bashbuild"
	"github.com/sipsma/bincastle/distro/builds/bcbuild"
	"github.com/sipsma/bincastle/distro/builds/binutilsbuild"
	"github.com/sipsma/bincastle/distro/builds/bisonbuild"
	"github.com/sipsma/bincastle/distro/builds/bzip2build"
	"github.com/sipsma/bincastle/distro/builds/cacertsbuild"
	"github.com/sipsma/bincastle/distro/builds/coreutilsbuild"
	"github.com/sipsma/bincastle/distro/builds/curlbuild"
	"github.com/sipsma/bincastle/distro/builds/diffutilsbuild"
	"github.com/sipsma/bincastle/distro/builds/e2fsprogsbuild"
	"github.com/sipsma/bincastle/distro/builds/elfutilsbuild"
	"github.com/sipsma/bincastle/distro/builds/expatbuild"
	"github.com/sipsma/bincastle/distro/builds/filebuild"
	"github.com/sipsma/bincastle/distro/builds/findutilsbuild"
	"github.com/sipsma/bincastle/distro/builds/flexbuild"
	"github.com/sipsma/bincastle/distro/builds/gccbuild"
	"github.com/sipsma/bincastle/distro/builds/gdbmbuild"
	"github.com/sipsma/bincastle/distro/builds/gettextbuild"
	"github.com/sipsma/bincastle/distro/builds/gitbuild"
	"github.com/sipsma/bincastle/distro/builds/gmpbuild"
	"github.com/sipsma/bincastle/distro/builds/golangbuild"
	"github.com/sipsma/bincastle/distro/builds/gperfbuild"
	"github.com/sipsma/bincastle/distro/builds/grepbuild"
	"github.com/sipsma/bincastle/distro/builds/groffbuild"
	"github.com/sipsma/bincastle/distro/builds/gzipbuild"
	"github.com/sipsma/bincastle/distro/builds/ianaetcbuild"
	"github.com/sipsma/bincastle/distro/builds/inetutilsbuild"
	"github.com/sipsma/bincastle/distro/builds/intltoolbuild"
	"github.com/sipsma/bincastle/distro/builds/iproute2build"
	"github.com/sipsma/bincastle/distro/builds/kbdbuild"
	"github.com/sipsma/bincastle/distro/builds/lessbuild"
	"github.com/sipsma/bincastle/distro/builds/libcapbuild"
	"github.com/sipsma/bincastle/distro/builds/libcbuild"
	"github.com/sipsma/bincastle/distro/builds/libffibuild"
	"github.com/sipsma/bincastle/distro/builds/libpipelinebuild"
	"github.com/sipsma/bincastle/distro/builds/libtasn1build"
	"github.com/sipsma/bincastle/distro/builds/libtoolbuild"
	"github.com/sipsma/bincastle/distro/builds/linuxbuild"
	"github.com/sipsma/bincastle/distro/builds/m4build"
	"github.com/sipsma/bincastle/distro/builds/makebuild"
	"github.com/sipsma/bincastle/distro/builds/mandbbuild"
	"github.com/sipsma/bincastle/distro/builds/manpagesbuild"
	"github.com/sipsma/bincastle/distro/builds/mesonbuild"
	"github.com/sipsma/bincastle/distro/builds/mpcbuild"
	"github.com/sipsma/bincastle/distro/builds/mpfrbuild"
	"github.com/sipsma/bincastle/distro/builds/ncursesbuild"
	"github.com/sipsma/bincastle/distro/builds/ninjabuild"
	"github.com/sipsma/bincastle/distro/builds/opensslbuild"
	"github.com/sipsma/bincastle/distro/builds/p11kitbuild"
	"github.com/sipsma/bincastle/distro/builds/patchbuild"
	"github.com/sipsma/bincastle/distro/builds/perl5build"
	"github.com/sipsma/bincastle/distro/builds/pkgconfigbuild"
	"github.com/sipsma/bincastle/distro/builds/procpsbuild"
	"github.com/sipsma/bincastle/distro/builds/psmiscbuild"
	"github.com/sipsma/bincastle/distro/builds/python3build"
	"github.com/sipsma/bincastle/distro/builds/readlinebuild"
	"github.com/sipsma/bincastle/distro/builds/sedbuild"
	"github.com/sipsma/bincastle/distro/builds/tarbuild"
	"github.com/sipsma/bincastle/distro/builds/texinfobuild"
	"github.com/sipsma/bincastle/distro/builds/usersbuild"
	"github.com/sipsma/bincastle/distro/builds/utillinuxbuild"
	"github.com/sipsma/bincastle/distro/builds/xzbuild"
	"github.com/sipsma/bincastle/distro/builds/zlibbuild"
	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/emacs"
	"github.com/sipsma/bincastle/distro/pkgs/autoconf"
	"github.com/sipsma/bincastle/distro/pkgs/automake"
	"github.com/sipsma/bincastle/distro/pkgs/awk"
	"github.com/sipsma/bincastle/distro/pkgs/bash"
	"github.com/sipsma/bincastle/distro/pkgs/bc"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/bison"
	"github.com/sipsma/bincastle/distro/pkgs/bzip2"
	"github.com/sipsma/bincastle/distro/pkgs/cacerts"
	"github.com/sipsma/bincastle/distro/pkgs/coreutils"
	"github.com/sipsma/bincastle/distro/pkgs/curl"
	"github.com/sipsma/bincastle/distro/pkgs/diffutils"
	"github.com/sipsma/bincastle/distro/pkgs/e2fsprogs"
	"github.com/sipsma/bincastle/distro/pkgs/elfutils"
	"github.com/sipsma/bincastle/distro/pkgs/expat"
	"github.com/sipsma/bincastle/distro/pkgs/file"
	"github.com/sipsma/bincastle/distro/pkgs/findutils"
	"github.com/sipsma/bincastle/distro/pkgs/flex"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gdbm"
	"github.com/sipsma/bincastle/distro/pkgs/gettext"
	"github.com/sipsma/bincastle/distro/pkgs/git"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/golang"
	"github.com/sipsma/bincastle/distro/pkgs/gperf"
	"github.com/sipsma/bincastle/distro/pkgs/grep"
	"github.com/sipsma/bincastle/distro/pkgs/groff"
	"github.com/sipsma/bincastle/distro/pkgs/gzip"
	"github.com/sipsma/bincastle/distro/pkgs/ianaetc"
	"github.com/sipsma/bincastle/distro/pkgs/inetutils"
	"github.com/sipsma/bincastle/distro/pkgs/intltool"
	"github.com/sipsma/bincastle/distro/pkgs/iproute2"
	"github.com/sipsma/bincastle/distro/pkgs/kbd"
	"github.com/sipsma/bincastle/distro/pkgs/less"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libcap"
	"github.com/sipsma/bincastle/distro/pkgs/libffi"
	"github.com/sipsma/bincastle/distro/pkgs/libpipeline"
	"github.com/sipsma/bincastle/distro/pkgs/libstdcpp"
	"github.com/sipsma/bincastle/distro/pkgs/libtasn1"
	"github.com/sipsma/bincastle/distro/pkgs/libtool"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	"github.com/sipsma/bincastle/distro/pkgs/make"
	"github.com/sipsma/bincastle/distro/pkgs/mandb"
	"github.com/sipsma/bincastle/distro/pkgs/manpages"
	"github.com/sipsma/bincastle/distro/pkgs/meson"
	"github.com/sipsma/bincastle/distro/pkgs/mpc"
	"github.com/sipsma/bincastle/distro/pkgs/mpfr"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/ninja"
	"github.com/sipsma/bincastle/distro/pkgs/openssl"
	"github.com/sipsma/bincastle/distro/pkgs/patch"
	"github.com/sipsma/bincastle/distro/pkgs/perl5"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/procps"
	"github.com/sipsma/bincastle/distro/pkgs/psmisc"
	"github.com/sipsma/bincastle/distro/pkgs/python3"
	"github.com/sipsma/bincastle/distro/pkgs/readline"
	"github.com/sipsma/bincastle/distro/pkgs/sed"
	"github.com/sipsma/bincastle/distro/pkgs/tar"
	"github.com/sipsma/bincastle/distro/pkgs/texinfo"
	"github.com/sipsma/bincastle/distro/pkgs/users"
	"github.com/sipsma/bincastle/distro/pkgs/utillinux"
	"github.com/sipsma/bincastle/distro/pkgs/xz"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	"github.com/sipsma/bincastle/distro/pkgs/p11kit"
)

type stage1Distro struct {
	Pkger
	distroSources
}

func (d stage1Distro) Binutils() binutils.Pkg {
	return binutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.BinutilsSrc(),
			),
			ScratchMount(`/build`),
			Shell(
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
				`mkdir -v /tools/lib`,
				`ln -sv lib /tools/lib64`,
				`make install`,
			),
		).With(
			Name("binutils-stage1"),
		)
	})
}

func (d stage1Distro) GCC() gcc.Pkg {
	return gcc.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.Binutils(),
				Patch(d, d.GCCSrc(), Shell(
					// TODO way of not having to hardcode "gcc-src" here.
					// Maybe have each Src pkg automatically AddEnv w/
					// the directory it's located at set to $<pkgname>_SRC_DIR or something
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
				)),
				d.MPFRSrc(),
				d.GMPSrc(),
				d.MPCSrc(),
			),
			ScratchMount(`/build`),
			Shell(
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
		).With(
			Name("gcc-stage1"),
			RuntimeDeps(d.Binutils()),
		)
	})
}

type stage2Distro struct {
	Pkger
	distroSources
}

func (d stage2Distro) LinuxHeaders() linux.HeadersPkg {
	return linux.BuildHeadersPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.LinuxSrc(), Shell(
					`cd /src/linux-src`,
					`make mrproper`,
				)).With(DiscardChanges()),
			),
			Shell(
				`cd /src/linux-src`,
				`make INSTALL_HDR_PATH=/tools headers_install`,
			),
		).With(
			Name("linux-headers-stage2"),
		)
	})
}

func (d stage2Distro) Libc() libc.Pkg {
	return libc.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LibcSrc(),
				d.LinuxHeaders(),
			),
			ScratchMount(`/build`),
			Shell(
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
		).With(
			Name("libc-stage2"),
		)
	})
}

func (d stage2Distro) Libstdcpp() libstdcpp.Pkg {
	return libstdcpp.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LibstdcppSrc(),
				d.Libc(),
				d.LinuxHeaders(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				`set -x`,
				`env`,
				`echo 'int main(){}' > dummy.c`,
				`x86_64-bincastle-linux-gnu-gcc dummy.c`,
				`readelf -l a.out | grep ': /tools'`,
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
		).With(
			Name("libstdcpp-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Binutils() binutils.Pkg {
	return binutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.BinutilsSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.Libstdcpp(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`CC=x86_64-bincastle-linux-gnu-gcc`,
					`AR=x86_64-bincastle-linux-gnu-ar`,
					`RANLIB=x86_64-bincastle-linux-gnu-ranlib`,
					`/src/binutils-src/configure`,
					`--prefix=/tools`,
					`--disable-nls`,
					`--disable-werror`,
					`--with-lib-path=/tools/lib`,
					`--with-sysroot`,
				}, " "),
				`make`,
				`make install`,
				`make -C ld clean`,
				`make -C ld LIB_PATH=/usr/lib:/lib`,
				`cp -v ld/ld-new /tools/bin`,
			),
		).With(
			Name("binutils-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) GCC() gcc.Pkg {
	return gcc.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.GCCSrc(), Shell(
					// TODO way of not having to hardcode "gcc-src" here.
					// Maybe have each Src pkg automatically AddEnv w/
					// the directory it's located at set to $<pkgname>_SRC_DIR or something
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
					`ln -s /src/mpfr-src /src/gcc-src/mpfr`,
					`ln -s /src/gmp-src /src/gcc-src/gmp`,
					`ln -s /src/mpc-src /src/gcc-src/mpc`,
				)),
				d.MPFRSrc(),
				d.GMPSrc(),
				d.MPCSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.Libstdcpp(),
				d.Binutils(),
			),
			ScratchMount(`/build`),
			Shell(
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
		).With(
			Name("gcc-stage2"),
			RuntimeDeps(
				d.Libc(),
				d.Libstdcpp(),
				d.Binutils(),
			),
		)
	})
}

func (d stage2Distro) M4() m4.Pkg {
	return m4.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.M4Src(), Shell(
					`cd /src/m4-src`,
					`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' lib/*.c`,
					`echo "#define _IO_IN_BACKUP 0x100" >> lib/stdio-impl.h`,
				)),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/m4-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("m4-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Ncurses() ncurses.Pkg {
	return ncurses.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.NcursesSrc(), Shell(
					`cd /src/ncurses-src`,
					`sed -i s/mawk// configure`,
				)),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/ncurses-src/configure`,
					`--prefix=/tools`,
					`--with-shared`,
					`--without-debug`,
					`--without-ada`,
					`--enable-widec`,
					`--enable-overwrite`,
				}, " "),
				`make`,
				`make install`,
				`ln -s libncursesw.so /tools/lib/libncurses.so`,
			),
		).With(
			Name("ncurses-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Bash() bash.Pkg {
	return bash.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.BashSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/bash-src/configure`,
					`--prefix=/tools`,
					`--without-bash-malloc`,
				}, " "),
				`make`,
				`make install`,
				`ln -sv bash /tools/bin/sh`,
			),
		).With(
			Name("bash-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Bison() bison.Pkg {
	return bison.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.BisonSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/bison-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make -j1`, // TODO
				`make install`,
			),
		).With(
			Name("bison-stage2"),
			RuntimeDeps(d.Libc(), d.M4()),
		)
	})
}

func (d stage2Distro) Bzip2() bzip2.Pkg {
	return bzip2.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.Bzip2Src().With(DiscardChanges()),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			Shell(
				`cd /src/bzip2-src`,
				`make`,
				`make PREFIX=/tools install`,
				`make clean`,
			),
		).With(
			Name("bzip2-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Coreutils() coreutils.Pkg {
	return coreutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.CoreutilsSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/coreutils-src/configure`,
					`--prefix=/tools`,
					`--enable-install-program=hostname`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("coreutils-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Diffutils() diffutils.Pkg {
	return diffutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.DiffutilsSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/diffutils-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("diffutils-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) File() file.Pkg {
	return file.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.FileSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/file-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("file-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Findutils() findutils.Pkg {
	return findutils.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.FindutilsSrc(), Shell(
					`cd /src/findutils-src`,
					`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c`,
					`sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c`,
					`echo "#define _IO_IN_BACKUP 0x100" >> gl/lib/stdio-impl.h`,
				)),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/findutils-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("findutils-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Awk() awk.Pkg {
	return awk.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.AwkSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/awk-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("awk-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Gettext() gettext.Pkg {
	return gettext.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.GettextSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
				d.Ncurses(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/gettext-src/configure`,
					`--disable-shared`,
				}, " "),
				`make`,
				`cp -v gettext-tools/src/{msgfmt,msgmerge,xgettext} /tools/bin`,
			),
		).With(
			Name("gettext-stage2"),
			RuntimeDeps(d.Libc(), d.Ncurses()),
		)
	})
}

func (d stage2Distro) Grep() grep.Pkg {
	return grep.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.GrepSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/grep-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("grep-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Gzip() gzip.Pkg {
	return gzip.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.GzipSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/gzip-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("gzip-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Make() make.Pkg {
	return make.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.MakeSrc(), Shell(
					`cd /src/make-src`,
					`sed -i '211,217 d; 219,229 d; 232 d' glob/glob.c`,
				)),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/make-src/configure`,
					`--prefix=/tools`,
					`--without-guile`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("make-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Patch() patch.Pkg {
	return patch.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.PatchSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/patch-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("patch-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Perl5() perl5.Pkg {
	return perl5.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.Perl5Src().With(DiscardChanges()),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
				d.Bzip2(),
			),
			Shell(
				`cd /src/perl5-src`,
				strings.Join([]string{`sh`, `Configure`,
					`-des`,
					`-Dprefix=/tools`,
					`-Dlibs=-lm`,
					`-Uloclibpth`,
					`-Ulocincpth`,
				}, " "),
				`make`,
				`cp -v perl cpan/podlators/scripts/pod2man /tools/bin`,
				`mkdir -pv /tools/lib/perl5/5.30.0`,
				`cp -Rv lib/* /tools/lib/perl5/5.30.0`,
				`make clean`,
			),
		).With(
			Name("perl5-stage2"),
			RuntimeDeps(d.Libc(), d.Bzip2()),
		)
	})
}

func (d stage2Distro) Python3() python3.Pkg {
	return python3.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				Patch(d, d.Python3Src(), Shell(
					`cd /src/python3-src`,
					`sed -i '/def add_multiarch_paths/a \        return' setup.py`,
				)),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
				d.Bzip2(),
				d.Ncurses(),
				d.File(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/python3-src/configure`,
					`--prefix=/tools`,
					`--without-ensurepip`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("python3-stage2"),
			RuntimeDeps(d.Libc(), d.Bzip2(), d.Ncurses(), d.File()),
		)
	})
}

func (d stage2Distro) Sed() sed.Pkg {
	return sed.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.SedSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/sed-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("sed-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Tar() tar.Pkg {
	return tar.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.TarSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/tar-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("tar-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

func (d stage2Distro) Texinfo() texinfo.Pkg {
	return texinfo.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.TexinfoSrc(),
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
				d.Perl5(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/texinfo-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("texinfo-stage2"),
			RuntimeDeps(d.Libc(), d.Perl5()),
		)
	})
}

func (d stage2Distro) Xz() xz.Pkg {
	return xz.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.GCC(),
				d.M4(),
				d.XzSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{`/src/xz-src/configure`,
					`--prefix=/tools`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("xz-stage2"),
			RuntimeDeps(d.Libc()),
		)
	})
}

type stage3Distro struct {
	Pkger
	distroSources
}

func (d stage3Distro) LinuxHeaders() linux.HeadersPkg {
	return linuxbuild.DefaultHeaders(d)
}

func (d stage3Distro) Libc() libc.Pkg {
	return libcbuild.DefaultGlibc(d)
}

type distro struct {
	Pkger
	stage3Distro
	distroSources
}

func (d distro) Manpages() manpages.Pkg {
	return manpagesbuild.Default(d)
}

func (d distro) Zlib() zlib.Pkg {
	return zlibbuild.Default(d)
}

func (d distro) File() file.Pkg {
	return filebuild.Default(d)
}

func (d distro) Readline() readline.Pkg {
	return readlinebuild.Default(d)
}

func (d distro) M4() m4.Pkg {
	return m4build.Default(d)
}

func (d distro) BC() bc.Pkg {
	return bcbuild.Default(d)
}

func (d distro) Binutils() binutils.Pkg {
	return binutilsbuild.Default(d)
}

func (d distro) GMP() gmp.Pkg {
	return gmpbuild.Default(d)
}

func (d distro) MPFR() mpfr.Pkg {
	return mpfrbuild.Default(d)
}

func (d distro) MPC() mpc.Pkg {
	return mpcbuild.Default(d)
}

func (d distro) GCC() gcc.Pkg {
	return gccbuild.Default(d)
}

func (d distro) Bzip2() bzip2.Pkg {
	return bzip2build.Default(d)
}

func (d distro) PkgConfig() pkgconfig.Pkg {
	return pkgconfigbuild.Default(d)
}

func (d distro) Ncurses() ncurses.Pkg {
	return ncursesbuild.Default(d)
}

func (d distro) Attr() attr.Pkg {
	return attrbuild.Default(d)
}

func (d distro) Acl() acl.Pkg {
	return aclbuild.Default(d)
}

func (d distro) Libcap() libcap.Pkg {
	return libcapbuild.Default(d)
}

func (d distro) Sed() sed.Pkg {
	return sedbuild.Default(d)
}

func (d distro) Psmisc() psmisc.Pkg {
	return psmiscbuild.Default(d)
}

func (d distro) Ianaetc() ianaetc.Pkg {
	return ianaetcbuild.Default(d)
}

func (d distro) Bison() bison.Pkg {
	return bisonbuild.Default(d)
}

func (d distro) Flex() flex.Pkg {
	return flexbuild.Default(d)
}

func (d distro) Grep() grep.Pkg {
	return grepbuild.Default(d)
}

func (d distro) Bash() bash.Pkg {
	return bashbuild.Default(d)
}

func (d distro) Libtool() libtool.Pkg {
	return libtoolbuild.Default(d)
}

func (d distro) GDBM() gdbm.Pkg {
	return gdbmbuild.Default(d)
}

func (d distro) Gperf() gperf.Pkg {
	return gperfbuild.Default(d)
}

func (d distro) Expat() expat.Pkg {
	return expatbuild.Default(d)
}

func (d distro) Inetutils() inetutils.Pkg {
	return inetutilsbuild.Default(d)
}

func (d distro) Perl5() perl5.Pkg {
	return perl5build.Default(d)
}

func (d distro) Perl5XMLParser() perl5.XMLParserPkg {
	return perl5build.DefaultXMLParser(d)
}

func (d distro) Intltool() intltool.Pkg {
	return intltoolbuild.Default(d)
}

func (d distro) Autoconf() autoconf.Pkg {
	return autoconfbuild.Default(d)
}

func (d distro) Automake() automake.Pkg {
	return automakebuild.Default(d)
}

func (d distro) Xz() xz.Pkg {
	return xzbuild.Default(d)
}

func (d distro) Gettext() gettext.Pkg {
	return gettextbuild.Default(d)
}

func (d distro) Elfutils() elfutils.Pkg {
	return elfutilsbuild.Default(d)
}

func (d distro) Libffi() libffi.Pkg {
	return libffibuild.Default(d)
}

func (d distro) OpenSSL() openssl.Pkg {
	return opensslbuild.Default(d)
}

func (d distro) Python3() python3.Pkg {
	return python3build.Default(d)
}

func (d distro) Ninja() ninja.Pkg {
	return ninjabuild.Default(d)
}

func (d distro) Meson() meson.Pkg {
	return mesonbuild.Default(d)
}

func (d distro) Coreutils() coreutils.Pkg {
	return coreutilsbuild.Default(d)
}

func (d distro) Diffutils() diffutils.Pkg {
	return diffutilsbuild.Default(d)
}

func (d distro) Awk() awk.Pkg {
	return awkbuild.Gawk(d)
}

func (d distro) Findutils() findutils.Pkg {
	return findutilsbuild.Default(d)
}

func (d distro) Groff() groff.Pkg {
	return groffbuild.Default(d)
}

func (d distro) Less() less.Pkg {
	return lessbuild.Default(d)
}

func (d distro) Gzip() gzip.Pkg {
	return gzipbuild.Default(d)
}

func (d distro) IPRoute2() iproute2.Pkg {
	return iproute2build.Default(d)
}

func (d distro) Kbd() kbd.Pkg {
	return kbdbuild.Default(d)
}

func (d distro) Libpipeline() libpipeline.Pkg {
	return libpipelinebuild.Default(d)
}

func (d distro) Make() make.Pkg {
	return makebuild.Default(d)
}

func (d distro) Patch() patch.Pkg {
	return patchbuild.Default(d)
}

func (d distro) ManDB() mandb.Pkg {
	return mandbbuild.Default(d)
}

func (d distro) Tar() tar.Pkg {
	return tarbuild.Default(d)
}

func (d distro) Texinfo() texinfo.Pkg {
	return texinfobuild.Default(d)
}

func (d distro) Procps() procps.Pkg {
	return procpsbuild.Default(d)
}

func (d distro) UtilLinux() utillinux.Pkg {
	return utillinuxbuild.Default(d)
}

func (d distro) E2fsprogs() e2fsprogs.Pkg {
	return e2fsprogsbuild.Default(d)
}

func (d distro) Libtasn1() libtasn1.Pkg {
	return libtasn1build.Default(d)
}

func (d distro) P11kit() p11kit.Pkg {
	return p11kitbuild.Default(d)
}

func (d distro) CACerts() cacerts.Pkg {
	return cacertsbuild.Default(d)
}

func (d distro) Curl() curl.Pkg {
	return curlbuild.Default(d)
}

func (d distro) Git() git.Pkg {
	return gitbuild.Default(d)
}

func (d distro) Golang() golang.Pkg {
	return golangbuild.Default(d)
}

func (d distro) Emacs() emacs.Pkg {
	return emacsbuild.Default(d)
}

func (d distro) Users() users.Pkg {
	return usersbuild.SingleUser(d,
		// TODO make this customizable by other consumers via a field in distro
		"sipsma",
		"/home/sipsma",
		"/bin/bash",
	)
}

// TODO hate this name...
func (d distro) MiscFiles() Pkg {
	return d.Exec(
		BuildDeps(
			d.Coreutils(),
			d.Bash(),
		),
		Shell(
			`ln -sv /run /var/run`,
			`mkdir -pv /var/{opt,cache,lib/{color,misc,locate},local}`,
			// TODO have to first remove the existing symlinks because
			// they are also provided in the bootstrap, which is later trimmed.
			// This is super ugly, you need to find a better way to de-dupe this
			// with the bootstrap
			`rm -f /bin/sh && ln -sv bash /bin/sh`,
			`rm -f /etc/mtab && ln -sv /proc/self/mounts /etc/mtab`,
		),
	).With(RuntimeDeps(d.Bash()))
}

func Bootstrap(bootstrapGraph Graph) Graph {
	defaultBuildOpts := []llb.RunOption{
		// TODO should this really be hardcoded?
		// maybe something smaller than "all cpus" would be a better default...
		llb.AddEnv("MAKEFLAGS", "-j"+strconv.Itoa(runtime.NumCPU())),
		llb.AddEnv("LC_ALL", "POSIX"),
		llb.AddEnv("FORCE_UNSAFE_CONFIGURE", "1"),
	}

	sources := distroSources{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/bin:/usr/bin"),
			BuildDeps(bootstrapGraph),
		)...),
	}

	tmpBaseSystem := DefaultPkger().Exec(
		BuildDeps(bootstrapGraph),
		Shell(
			`ln -sv /sysroot/tools /`,
			`mkdir -pv /sysroot/tools`,
		),
	).With(RuntimeDeps(bootstrapGraph), Name("tmpBaseSystem"))

	stage1 := func(d stage1Distro) Graph {
		return Merge(
			d.Binutils(),
			d.GCC(),
		)
	}(stage1Distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
			BuildDeps(tmpBaseSystem),
			AtRuntime(
				RuntimeDeps(tmpBaseSystem),
			),
		)...),
		distroSources: sources,
	})

	stage2 := func(d stage2Distro) Graph {
		return Merge(
			d.LinuxHeaders(),
			d.Libc(),
			d.Binutils(),
			d.GCC(),
			d.M4(),
			d.Ncurses(),
			d.Bash(),
			d.Bison(),
			d.Bzip2(),
			d.Coreutils(),
			d.Diffutils(),
			d.File(),
			d.Findutils(),
			d.Awk(),
			d.Gettext(),
			d.Make(),
			d.Grep(),
			d.Gzip(),
			d.Patch(),
			d.Perl5(),
			d.Python3(),
			d.Sed(),
			d.Tar(),
			d.Texinfo(),
			d.Xz(),
		)
	}(stage2Distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
			BuildDeps(stage1),
			AtRuntime(
				RuntimeDeps(stage1),
			),
		)...),
		distroSources: sources,
	})

	unpatchedTmp := Transform(
		TrimGraphs(stage2, bootstrapGraph), OutputDir("/sysroot"),
	)

	baseSystem := DefaultPkger().Exec(
		BuildDeps(unpatchedTmp),
		llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
		Shell(
			`mkdir -pv /{bin,boot,etc/{opt,sysconfig},home,lib/firmware,mnt,opt}`,
			`mkdir -pv /{media/{floppy,cdrom},sbin,srv,var}`,
			`install -dv -m 0750 /root`,
			`install -dv -m 1777 /tmp /var/tmp`,
			`mkdir -pv /usr/{,local/}{bin,include,lib,sbin,src}`,
			`mkdir -pv /usr/{,local/}share/{color,dict,doc,info,locale,man}`,
			`mkdir -v  /usr/{,local/}share/{misc,terminfo,zoneinfo}`,
			`mkdir -v  /usr/libexec`,
			`mkdir -pv /usr/{,local/}share/man/man{1..8}`,
			`mkdir -v  /usr/lib/pkgconfig`,
			`mkdir -v /lib64`,
			`mkdir -v /var/{log,mail,spool}`,
			`ln -sv /run /var/run`,
			`ln -sv /run/lock /var/lock`,
			`mkdir -pv /var/{opt,cache,lib/{color,misc,locate},local}`,
			`ln -sv /tools/bin/{bash,cat,chmod,dd,echo,ln,mkdir,pwd,rm,stty,touch} /bin`,
			`ln -sv /tools/bin/{env,install,perl,printf} /usr/bin`,
			`ln -sv /tools/lib/libgcc_s.so{,.1} /usr/lib`,
			`ln -sv /tools/lib/libstdc++.{a,so{,.6}} /usr/lib`,
			`ln -sv bash /bin/sh`,
			`ln -sv /proc/self/mounts /etc/mtab`,
		),
	).With(RuntimeDeps(unpatchedTmp), Name("base-system"))

	stage3 := stage3Distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin:/tools/bin"),
			BuildDeps(baseSystem),
			AtRuntime(RuntimeDeps(baseSystem)),
		)...),
		distroSources: sources,
	}

	patchedTmp := DefaultPkger().Exec(
		BuildDeps(baseSystem),
		llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
		Shell(
			`mv -v /tools/bin/{ld,ld-old}`,
			`mv -v /tools/$(uname -m)-pc-linux-gnu/bin/{ld,ld-old}`,
			`mv -v /tools/bin/{ld-new,ld}`,
			`ln -sv /tools/bin/ld /tools/$(uname -m)-pc-linux-gnu/bin/ld`,
			`gcc -dumpspecs | sed -e 's@/tools@@g' -e '/\*startfile_prefix_spec:/{n;s@.*@/usr/lib/ @}'  -e '/\*cpp:/{n;s@$@ -isystem /usr/include@}' > $(dirname $(gcc --print-libgcc-file-name))/specs`,
		),
	).With(RuntimeDeps(baseSystem), Name("patched-stage2"))

	distroGraph := func(d distro) Graph {
		return Merge(
			d.Libc(),
			d.LinuxHeaders(),
			d.Manpages(),
			d.Zlib(),
			d.File(),
			d.Readline(),
			d.M4(),
			d.BC(),
			d.Binutils(),
			d.GMP(),
			d.MPFR(),
			d.MPC(),
			d.GCC(),
			d.Bzip2(),
			d.PkgConfig(),
			d.Ncurses(),
			d.Attr(),
			d.Acl(),
			d.Libcap(),
			d.Sed(),
			d.Psmisc(),
			d.Ianaetc(),
			d.Bison(),
			d.Flex(),
			d.Grep(),
			d.Bash(),
			d.Libtool(),
			d.GDBM(),
			d.Gperf(),
			d.Expat(),
			d.Inetutils(),
			d.Perl5(),
			d.Perl5XMLParser(),
			d.Intltool(),
			d.Autoconf(),
			d.Automake(),
			d.Xz(),
			d.Gettext(),
			d.Elfutils(),
			d.Libffi(),
			d.OpenSSL(),
			d.Python3(),
			d.Ninja(),
			d.Meson(),
			d.Coreutils(),
			d.Diffutils(),
			d.Awk(),
			d.Findutils(),
			d.Groff(),
			d.Less(),
			d.Gzip(),
			d.IPRoute2(),
			d.Kbd(),
			d.Libpipeline(),
			d.Make(),
			d.Patch(),
			d.ManDB(),
			d.Tar(),
			d.Texinfo(),
			d.Xz(),
			d.Procps(),
			d.UtilLinux(),
			d.E2fsprogs(),
			d.Libtasn1(),
			d.CACerts(),
			d.Curl(),
			d.Git(),
			d.Golang(),
			d.Users(),
			d.MiscFiles(),
		)
	}(distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin:/tools/bin"),
			BuildDeps(patchedTmp),
			AtRuntime(RuntimeDeps(patchedTmp)),
		)...),
		stage3Distro:  stage3,
		distroSources: sources,
	})

	return TrimGraphs(distroGraph, patchedTmp, baseSystem)
}
