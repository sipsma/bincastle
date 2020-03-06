package distro

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/moby/buildkit/client/llb"

	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"

	"github.com/sipsma/bincastle/distro/builds/aclbuild"
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
	"github.com/sipsma/bincastle/distro/builds/utillinuxbuild"
	"github.com/sipsma/bincastle/distro/builds/xzbuild"
	"github.com/sipsma/bincastle/distro/builds/zlibbuild"
	"github.com/sipsma/bincastle/distro/builds/usersbuild"
	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/users"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
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
	"github.com/sipsma/bincastle/distro/pkgs/utillinux"
	"github.com/sipsma/bincastle/distro/pkgs/xz"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
)

type stage1Distro struct {
	Pkger
	distroSources
}

func (d stage1Distro) Binutils() PkgBuild {
	return PkgBuildOf(d.Exec(
		binutils.SrcPkg(d),
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
		VersionOf(binutils.SrcPkg(d)),
	))
}

func (d stage1Distro) GCC() PkgBuild {
	return PkgBuildOf(d.Exec(
		binutils.Pkg(d),
		Patch(d, gcc.SrcPkg(d), Shell(
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
		mpfr.SrcPkg(d),
		gmp.SrcPkg(d),
		mpc.SrcPkg(d),
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
		// TODO remove
		AtRuntime(Name("gcc-stage1")),
	).With(
		Name("gcc-stage1"),
		Deps(binutils.Pkg(d)),
		VersionOf(gcc.SrcPkg(d)),
	))
}

type stage2Distro struct {
	Pkger
	distroSources
}

func (d stage2Distro) LinuxHeaders() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, linux.SrcPkg(d), Shell(
			`cd /src/linux-src`,
			`make mrproper`,
		)).With(DiscardChanges()),
		Shell(
			`cd /src/linux-src`,
			`make INSTALL_HDR_PATH=/tools headers_install`,
		),
	).With(
		Name("linux-headers-stage2"),
		VersionOf(linux.SrcPkg(d)),
	))
}

func (d stage2Distro) Libc() PkgBuild {
	return PkgBuildOf(d.Exec(
		libc.SrcPkg(d),
		linux.HeadersPkg(d),
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
		VersionOf(libc.SrcPkg(d)),
	))
}

func (d stage2Distro) Libstdcpp() PkgBuild {
	return PkgBuildOf(d.Exec(
		libstdcpp.SrcPkg(d),
		libc.Pkg(d),
		linux.HeadersPkg(d),
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
		VersionOf(libstdcpp.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Binutils() PkgBuild {
	return PkgBuildOf(d.Exec(
		binutils.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		libstdcpp.Pkg(d),
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
		VersionOf(binutils.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) GCC() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, gcc.SrcPkg(d), Shell(
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
		mpfr.SrcPkg(d),
		gmp.SrcPkg(d),
		mpc.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		libstdcpp.Pkg(d),
		binutils.Pkg(d),
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
		VersionOf(gcc.SrcPkg(d)),
		Deps(
			libc.Pkg(d),
			libstdcpp.Pkg(d),
			binutils.Pkg(d),
		),
	))
}

func (d stage2Distro) M4() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, m4.SrcPkg(d), Shell(
			`cd /src/m4-src`,
			`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' lib/*.c`,
			`echo "#define _IO_IN_BACKUP 0x100" >> lib/stdio-impl.h`,
		)),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
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
		VersionOf(m4.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Ncurses() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, ncurses.SrcPkg(d), Shell(
			`cd /src/ncurses-src`,
			`sed -i s/mawk// configure`,
		)),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(ncurses.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Bash() PkgBuild {
	return PkgBuildOf(d.Exec(
		bash.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(bash.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Bison() PkgBuild {
	return PkgBuildOf(d.Exec(
		bison.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(bison.SrcPkg(d)),
		Deps(libc.Pkg(d), m4.Pkg(d)),
	))
}

func (d stage2Distro) Bzip2() PkgBuild {
	return PkgBuildOf(d.Exec(
		bzip2.SrcPkg(d).With(DiscardChanges()),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		Shell(
			`cd /src/bzip2-src`,
			`make`,
			`make PREFIX=/tools install`,
			`make clean`,
		),
	).With(
		Name("bzip2-stage2"),
		VersionOf(bzip2.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Coreutils() PkgBuild {
	return PkgBuildOf(d.Exec(
		coreutils.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(coreutils.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Diffutils() PkgBuild {
	return PkgBuildOf(d.Exec(
		diffutils.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(diffutils.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) File() PkgBuild {
	return PkgBuildOf(d.Exec(
		file.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(file.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Findutils() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, findutils.SrcPkg(d), Shell(
			`cd /src/findutils-src`,
			`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' gl/lib/*.c`,
			`sed -i '/unistd/a #include <sys/sysmacros.h>' gl/lib/mountlist.c`,
			`echo "#define _IO_IN_BACKUP 0x100" >> gl/lib/stdio-impl.h`,
		)),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(findutils.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Awk() PkgBuild {
	return PkgBuildOf(d.Exec(
		awk.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(awk.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Gettext() PkgBuild {
	return PkgBuildOf(d.Exec(
		gettext.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		ncurses.Pkg(d),
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
		VersionOf(gettext.SrcPkg(d)),
		Deps(libc.Pkg(d), ncurses.Pkg(d)),
	))
}

func (d stage2Distro) Grep() PkgBuild {
	return PkgBuildOf(d.Exec(
		grep.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(grep.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Gzip() PkgBuild {
	return PkgBuildOf(d.Exec(
		gzip.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(gzip.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Make() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, make.SrcPkg(d), Shell(
			`cd /src/make-src`,
			`sed -i '211,217 d; 219,229 d; 232 d' glob/glob.c`,
		)),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(make.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Patch() PkgBuild {
	return PkgBuildOf(d.Exec(
		patch.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(patch.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Perl5() PkgBuild {
	return PkgBuildOf(d.Exec(
		perl5.SrcPkg(d).With(DiscardChanges()),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		bzip2.Pkg(d),
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
		VersionOf(perl5.SrcPkg(d)),
		Deps(libc.Pkg(d), bzip2.Pkg(d)),
	))
}

func (d stage2Distro) Python3() PkgBuild {
	return PkgBuildOf(d.Exec(
		Patch(d, python3.SrcPkg(d), Shell(
			`cd /src/python3-src`,
			`sed -i '/def add_multiarch_paths/a \        return' setup.py`,
		)),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		bzip2.Pkg(d),
		ncurses.Pkg(d),
		file.Pkg(d),
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
		VersionOf(python3.SrcPkg(d)),
		Deps(libc.Pkg(d), bzip2.Pkg(d), ncurses.Pkg(d), file.Pkg(d)),
	))
}

func (d stage2Distro) Sed() PkgBuild {
	return PkgBuildOf(d.Exec(
		sed.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(sed.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Tar() PkgBuild {
	return PkgBuildOf(d.Exec(
		tar.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(tar.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

func (d stage2Distro) Texinfo() PkgBuild {
	return PkgBuildOf(d.Exec(
		texinfo.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
		perl5.Pkg(d),
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
		VersionOf(texinfo.SrcPkg(d)),
		Deps(libc.Pkg(d), perl5.Pkg(d)),
	))
}

func (d stage2Distro) Xz() PkgBuild {
	return PkgBuildOf(d.Exec(
		xz.SrcPkg(d),
		linux.HeadersPkg(d),
		libc.Pkg(d),
		gcc.Pkg(d),
		m4.Pkg(d),
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
		VersionOf(xz.SrcPkg(d)),
		Deps(libc.Pkg(d)),
	))
}

type stage3Distro struct {
	Pkger
	distroSources
}

func (d stage3Distro) LinuxHeaders() PkgBuild {
	return linuxbuild.DefaultHeaders(d)
}

func (d stage3Distro) Libc() PkgBuild {
	return libcbuild.DefaultGlibc(d)
}

type distro struct {
	Pkger
	stage3Distro
	distroSources
}

func (d distro) Manpages() PkgBuild {
	return manpagesbuild.Default(d)
}

func (d distro) Zlib() PkgBuild {
	return zlibbuild.Default(d)
}

func (d distro) File() PkgBuild {
	return filebuild.Default(d)
}

func (d distro) Readline() PkgBuild {
	return readlinebuild.Default(d)
}

func (d distro) M4() PkgBuild {
	return m4build.Default(d)
}

func (d distro) BC() PkgBuild {
	return bcbuild.Default(d)
}

func (d distro) Binutils() PkgBuild {
	return binutilsbuild.Default(d)
}

func (d distro) GMP() PkgBuild {
	return gmpbuild.Default(d)
}

func (d distro) MPFR() PkgBuild {
	return mpfrbuild.Default(d)
}

func (d distro) MPC() PkgBuild {
	return mpcbuild.Default(d)
}

func (d distro) GCC() PkgBuild {
	return gccbuild.Default(d)
}

func (d distro) Bzip2() PkgBuild {
	return bzip2build.Default(d)
}

func (d distro) PkgConfig() PkgBuild {
	return pkgconfigbuild.Default(d)
}

func (d distro) Ncurses() PkgBuild {
	return ncursesbuild.Default(d)
}

func (d distro) Attr() PkgBuild {
	return attrbuild.Default(d)
}

func (d distro) Acl() PkgBuild {
	return aclbuild.Default(d)
}

func (d distro) Libcap() PkgBuild {
	return libcapbuild.Default(d)
}

func (d distro) Sed() PkgBuild {
	return sedbuild.Default(d)
}

func (d distro) Psmisc() PkgBuild {
	return psmiscbuild.Default(d)
}

func (d distro) Ianaetc() PkgBuild {
	return ianaetcbuild.Default(d)
}

func (d distro) Bison() PkgBuild {
	return bisonbuild.Default(d)
}

func (d distro) Flex() PkgBuild {
	return flexbuild.Default(d)
}

func (d distro) Grep() PkgBuild {
	return grepbuild.Default(d)
}

func (d distro) Bash() PkgBuild {
	return bashbuild.Default(d)
}

func (d distro) Libtool() PkgBuild {
	return libtoolbuild.Default(d)
}

func (d distro) GDBM() PkgBuild {
	return gdbmbuild.Default(d)
}

func (d distro) Gperf() PkgBuild {
	return gperfbuild.Default(d)
}

func (d distro) Expat() PkgBuild {
	return expatbuild.Default(d)
}

func (d distro) Inetutils() PkgBuild {
	return inetutilsbuild.Default(d)
}

func (d distro) Perl5() PkgBuild {
	return perl5build.Default(d)
}

func (d distro) Perl5XMLParser() PkgBuild {
	return perl5build.DefaultXMLParser(d)
}

func (d distro) Intltool() PkgBuild {
	return intltoolbuild.Default(d)
}

func (d distro) Autoconf() PkgBuild {
	return autoconfbuild.Default(d)
}

func (d distro) Automake() PkgBuild {
	return automakebuild.Default(d)
}

func (d distro) Xz() PkgBuild {
	return xzbuild.Default(d)
}

func (d distro) Gettext() PkgBuild {
	return gettextbuild.Default(d)
}

func (d distro) Elfutils() PkgBuild {
	return elfutilsbuild.Default(d)
}

func (d distro) Libffi() PkgBuild {
	return libffibuild.Default(d)
}

func (d distro) OpenSSL() PkgBuild {
	return opensslbuild.Default(d)
}

func (d distro) Python3() PkgBuild {
	return python3build.Default(d)
}

func (d distro) Ninja() PkgBuild {
	return ninjabuild.Default(d)
}

func (d distro) Meson() PkgBuild {
	return mesonbuild.Default(d)
}

func (d distro) Coreutils() PkgBuild {
	return coreutilsbuild.Default(d)
}

func (d distro) Diffutils() PkgBuild {
	return diffutilsbuild.Default(d)
}

func (d distro) Awk() PkgBuild {
	return awkbuild.Gawk(d)
}

func (d distro) Findutils() PkgBuild {
	return findutilsbuild.Default(d)
}

func (d distro) Groff() PkgBuild {
	return groffbuild.Default(d)
}

func (d distro) Less() PkgBuild {
	return lessbuild.Default(d)
}

func (d distro) Gzip() PkgBuild {
	return gzipbuild.Default(d)
}

func (d distro) IPRoute2() PkgBuild {
	return iproute2build.Default(d)
}

func (d distro) Kbd() PkgBuild {
	return kbdbuild.Default(d)
}

func (d distro) Libpipeline() PkgBuild {
	return libpipelinebuild.Default(d)
}

func (d distro) Make() PkgBuild {
	return makebuild.Default(d)
}

func (d distro) Patch() PkgBuild {
	return patchbuild.Default(d)
}

func (d distro) ManDB() PkgBuild {
	return mandbbuild.Default(d)
}

func (d distro) Tar() PkgBuild {
	return tarbuild.Default(d)
}

func (d distro) Texinfo() PkgBuild {
	return texinfobuild.Default(d)
}

func (d distro) Procps() PkgBuild {
	return procpsbuild.Default(d)
}

func (d distro) UtilLinux() PkgBuild {
	return utillinuxbuild.Default(d)
}

func (d distro) E2fsprogs() PkgBuild {
	return e2fsprogsbuild.Default(d)
}

func (d distro) Libtasn1() PkgBuild {
	return libtasn1build.Default(d)
}

func (d distro) P11kit() PkgBuild {
	return p11kitbuild.Default(d)
}

func (d distro) CACerts() PkgBuild {
	return cacertsbuild.Default(d)
}

func (d distro) Curl() PkgBuild {
	return curlbuild.Default(d)
}

func (d distro) Git() PkgBuild {
	return gitbuild.Default(d)
}

/* TODO
func (d distro) Golang() PkgBuild {
	return golangbuild.Default(d)
}
*/

func (d distro) Users() PkgBuild {
	return usersbuild.SingleUser(d,
		// TODO make this customizable by other consumers via a field in distro
		"sipsma",
		"/home/sipsma",
		"/bin/bash",
	)
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
			bootstrapGraph,
		)...),
	}

	tmpBaseSystem := DefaultPkger().Exec(
		bootstrapGraph,
		Shell(
			`ln -sv /sysroot/tools /`,
			`mkdir -pv /sysroot/tools`,
		),
	).With(Deps(bootstrapGraph), Name("tmpBaseSystem"))

	stage1 := func(d stage1Distro) Graph {
		return Merge(
			binutils.Pkg(d),
			gcc.Pkg(d),
		)
	}(stage1Distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
			tmpBaseSystem,
			AtRuntime(
				Deps(tmpBaseSystem),
			),
		)...),
		distroSources: sources,
	})

	stage2 := func(d stage2Distro) Graph {
		return Merge(
			linux.HeadersPkg(d),
			libc.Pkg(d),
			binutils.Pkg(d),
			gcc.Pkg(d),
			m4.Pkg(d),
			ncurses.Pkg(d),
			bash.Pkg(d),
			bison.Pkg(d),
			bzip2.Pkg(d),
			coreutils.Pkg(d),
			diffutils.Pkg(d),
			file.Pkg(d),
			findutils.Pkg(d),
			awk.Pkg(d),
			gettext.Pkg(d),
			grep.Pkg(d),
			gzip.Pkg(d),
			make.Pkg(d),
			patch.Pkg(d),
			perl5.Pkg(d),
			python3.Pkg(d),
			sed.Pkg(d),
			tar.Pkg(d),
			texinfo.Pkg(d),
			xz.Pkg(d),
		)
	}(stage2Distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
			stage1,
			AtRuntime(
				Deps(stage1),
			),
		)...),
		distroSources: sources,
	})

	unpatchedTmp := Transform(
		TrimGraphs(stage2, bootstrapGraph), OutputDir("/sysroot"),
	)

	baseSystem := DefaultPkger().Exec(
		unpatchedTmp,
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
	).With(Deps(unpatchedTmp), Name("base-system"))

	stage3 := stage3Distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin:/tools/bin"),
			baseSystem,
			AtRuntime(Deps(baseSystem)),
		)...),
		distroSources: sources,
	}

	patchedTmp := DefaultPkger().Exec(
		baseSystem,
		llb.AddEnv("PATH", "/tools/bin:/bin:/usr/bin"),
		Shell(
			`mv -v /tools/bin/{ld,ld-old}`,
			`mv -v /tools/$(uname -m)-pc-linux-gnu/bin/{ld,ld-old}`,
			`mv -v /tools/bin/{ld-new,ld}`,
			`ln -sv /tools/bin/ld /tools/$(uname -m)-pc-linux-gnu/bin/ld`,
			`gcc -dumpspecs | sed -e 's@/tools@@g' -e '/\*startfile_prefix_spec:/{n;s@.*@/usr/lib/ @}'  -e '/\*cpp:/{n;s@$@ -isystem /usr/include@}' > $(dirname $(gcc --print-libgcc-file-name))/specs`,
		),
	).With(Deps(baseSystem), Name("patched-stage2"))

	distroGraph := func(d distro) Graph {
		return Merge(
			libc.Pkg(d),
			linux.HeadersPkg(d),
			manpages.Pkg(d),
			zlib.Pkg(d),
			file.Pkg(d),
			readline.Pkg(d),
			m4.Pkg(d),
			bc.Pkg(d),
			binutils.Pkg(d),
			gmp.Pkg(d),
			mpfr.Pkg(d),
			mpc.Pkg(d),
			gcc.Pkg(d),
			bzip2.Pkg(d),
			pkgconfig.Pkg(d),
			ncurses.Pkg(d),
			attr.Pkg(d),
			acl.Pkg(d),
			libcap.Pkg(d),
			sed.Pkg(d),
			psmisc.Pkg(d),
			ianaetc.Pkg(d),
			bison.Pkg(d),
			flex.Pkg(d),
			grep.Pkg(d),
			bash.Pkg(d),
			libtool.Pkg(d),
			gdbm.Pkg(d),
			gperf.Pkg(d),
			expat.Pkg(d),
			inetutils.Pkg(d),
			perl5.Pkg(d),
			perl5.XMLParserPkg(d),
			intltool.Pkg(d),
			autoconf.Pkg(d),
			automake.Pkg(d),
			xz.Pkg(d),
			gettext.Pkg(d),
			elfutils.Pkg(d),
			libffi.Pkg(d),
			openssl.Pkg(d),
			python3.Pkg(d),
			ninja.Pkg(d),
			meson.Pkg(d),
			coreutils.Pkg(d),
			diffutils.Pkg(d),
			awk.Pkg(d),
			findutils.Pkg(d),
			groff.Pkg(d),
			less.Pkg(d),
			gzip.Pkg(d),
			iproute2.Pkg(d),
			kbd.Pkg(d),
			libpipeline.Pkg(d),
			make.Pkg(d),
			patch.Pkg(d),
			mandb.Pkg(d),
			tar.Pkg(d),
			texinfo.Pkg(d),
			procps.Pkg(d),
			utillinux.Pkg(d),
			e2fsprogs.Pkg(d),
			libtasn1.Pkg(d),
			cacerts.Pkg(d),
			curl.Pkg(d),
			git.Pkg(d),
			users.Pkg(d),
		)
	}(distro{
		Pkger: DefaultPkger(append(defaultBuildOpts,
			llb.AddEnv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin:/tools/bin"),
			patchedTmp,
			AtRuntime(Deps(patchedTmp)),
		)...),
		stage3Distro:  stage3,
		distroSources: sources,
	})

	return TrimGraphs(distroGraph, patchedTmp, baseSystem)
}
