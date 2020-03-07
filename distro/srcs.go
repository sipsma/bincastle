package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"

	"github.com/sipsma/bincastle/distro/pkgs/golang"
)

type distroSources struct {
	Pkger
}

func (d distroSources) AwkSrc() PkgBuild {
	return src.Curl(d,
		Name("awk-src"),
		src.URL("https://ftp.gnu.org/gnu/gawk/gawk-5.0.1.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) BashSrc() PkgBuild {
	return src.Curl(d,
		Name("bash-src"),
		src.URL("https://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) BinutilsSrc() PkgBuild {
	return src.Curl(d,
		Name("binutils-src"),
		src.URL("https://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) BisonSrc() PkgBuild {
	return src.Curl(d,
		Name("bison-src"),
		src.URL("https://ftp.gnu.org/gnu/bison/bison-3.4.1.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) Bzip2Src() PkgBuild {
	return src.Curl(d,
		Name("bzip2-src"),
		src.URL("https://www.sourceware.org/pub/bzip2/bzip2-1.0.8.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) CoreutilsSrc() PkgBuild {
	return src.Curl(d,
		Name("coreutils-src"),
		src.URL("https://ftp.gnu.org/gnu/coreutils/coreutils-8.31.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) DiffutilsSrc() PkgBuild {
	return src.Curl(d,
		Name("diffutils-src"),
		src.URL("https://ftp.gnu.org/gnu/diffutils/diffutils-3.7.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) FileSrc() PkgBuild {
	return src.Curl(d,
		Name("file-src"),
		src.URL("ftp://ftp.astron.com/pub/file/file-5.37.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) FindutilsSrc() PkgBuild {
	return src.Curl(d,
		Name("findutils-src"),
		src.URL("https://ftp.gnu.org/gnu/findutils/findutils-4.6.0.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GCCSrc() PkgBuild {
	return src.Curl(d,
		Name("gcc-src"),
		src.URL("http://ftp.gnu.org/gnu/gcc/gcc-9.2.0/gcc-9.2.0.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GettextSrc() PkgBuild {
	return src.Curl(d,
		Name("gettext-src"),
		src.URL("https://ftp.gnu.org/gnu/gettext/gettext-0.20.1.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GMPSrc() PkgBuild {
	return src.Curl(d,
		Name("gmp-src"),
		src.URL("https://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GrepSrc() PkgBuild {
	return src.Curl(d,
		Name("grep-src"),
		src.URL("https://ftp.gnu.org/gnu/grep/grep-3.3.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GzipSrc() PkgBuild {
	return src.Curl(d,
		Name("gzip-src"),
		src.URL("https://ftp.gnu.org/gnu/gzip/gzip-1.10.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LibcSrc() PkgBuild {
	return src.Curl(d,
		Name("libc-src"),
		src.URL("http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LibstdcppSrc() PkgBuild {
	// TODO you should change the outputdir to be down a
	// level at ./libstdc++. This will require changing
	// OutputDir to append if provided a relpath.
	return d.GCCSrc()
}

func (d distroSources) LinuxSrc() PkgBuild {
	return src.Curl(d,
		Name("linux-src"),
		src.URL("https://www.kernel.org/pub/linux/kernel/v5.x/linux-5.2.8.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) M4Src() PkgBuild {
	return src.Curl(d,
		Name("m4-src"),
		src.URL("https://ftp.gnu.org/gnu/m4/m4-1.4.18.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) MakeSrc() PkgBuild {
	return src.Curl(d,
		Name("make-src"),
		src.URL("https://ftp.gnu.org/gnu/make/make-4.2.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) MPCSrc() PkgBuild {
	return src.Curl(d,
		Name("mpc-src"),
		src.URL("https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) MPFRSrc() PkgBuild {
	return src.Curl(d,
		Name("mpfr-src"),
		src.URL("https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) NcursesSrc() PkgBuild {
	return src.Curl(d,
		Name("ncurses-src"),
		src.URL("https://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) PatchSrc() PkgBuild {
	return src.Curl(d,
		Name("patch-src"),
		src.URL("https://ftp.gnu.org/gnu/patch/patch-2.7.6.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) Perl5Src() PkgBuild {
	return src.Curl(d,
		Name("perl5-src"),
		src.URL("https://www.cpan.org/src/5.0/perl-5.30.0.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) Python3Src() PkgBuild {
	return src.Curl(d,
		Name("python3-src"),
		src.URL("https://www.python.org/ftp/python/3.7.4/Python-3.7.4.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) SedSrc() PkgBuild {
	return src.Curl(d,
		Name("sed-src"),
		src.URL("https://ftp.gnu.org/gnu/sed/sed-4.7.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) TarSrc() PkgBuild {
	return src.Curl(d,
		Name("tar-src"),
		src.URL("https://ftp.gnu.org/gnu/tar/tar-1.32.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) TexinfoSrc() PkgBuild {
	return src.Curl(d,
		Name("texinfo-src"),
		src.URL("https://ftp.gnu.org/gnu/texinfo/texinfo-6.6.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) XzSrc() PkgBuild {
	return src.Curl(d,
		Name("xz-src"),
		src.URL("https://astuteinternet.dl.sourceforge.net/project/lzmautils/xz-5.2.4.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) AclSrc() PkgBuild {
	return src.Curl(d,
		Name("acl-src"),
		src.URL("http://download.savannah.gnu.org/releases/acl/acl-2.2.53.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) AttrSrc() PkgBuild {
	return src.Curl(d,
		Name("attr-src"),
		src.URL("http://download.savannah.gnu.org/releases/attr/attr-2.4.48.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) AutoconfSrc() PkgBuild {
	return src.Curl(d,
		Name("autoconf-src"),
		src.URL("http://ftp.gnu.org/gnu/autoconf/autoconf-2.69.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) AutomakeSrc() PkgBuild {
	return src.Curl(d,
		Name("automake-src"),
		src.URL("http://ftp.gnu.org/gnu/automake/automake-1.16.1.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) BCSrc() PkgBuild {
	return src.Curl(d,
		Name("bc-src"),
		src.URL("https://github.com/gavinhoward/bc/archive/2.1.3/bc-2.1.3.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) E2fsprogsSrc() PkgBuild {
	return src.Curl(d,
		Name("e2fsprogs-src"),
		src.URL("https://downloads.sourceforge.net/project/e2fsprogs/e2fsprogs/v1.45.3/e2fsprogs-1.45.3.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) ElfutilsSrc() PkgBuild {
	return src.Curl(d,
		Name("elfutils-src"),
		src.URL("https://sourceware.org/ftp/elfutils/0.177/elfutils-0.177.tar.bz2"),
		src.StripComponents(1),
	)
}

func (d distroSources) ExpatSrc() PkgBuild {
	return src.Curl(d,
		Name("expat-src"),
		src.URL("https://prdownloads.sourceforge.net/expat/expat-2.2.7.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) FlexSrc() PkgBuild {
	return src.Curl(d,
		Name("flex-src"),
		src.URL("https://github.com/westes/flex/releases/download/v2.6.4/flex-2.6.4.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GDBMSrc() PkgBuild {
	return src.Curl(d,
		Name("gdbm-src"),
		src.URL("http://ftp.gnu.org/gnu/gdbm/gdbm-1.18.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GperfSrc() PkgBuild {
	return src.Curl(d,
		Name("gperf-src"),
		src.URL("http://ftp.gnu.org/gnu/gperf/gperf-3.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GroffSrc() PkgBuild {
	return src.Curl(d,
		Name("groff-src"),
		src.URL("http://ftp.gnu.org/gnu/groff/groff-1.22.4.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) IanaetcSrc() PkgBuild {
	return src.Curl(d,
		Name("ianaetc-src"),
		src.URL("http://anduin.linuxfromscratch.org/LFS/iana-etc-2.30.tar.bz2"),
		src.StripComponents(1),
	)
}

func (d distroSources) InetutilsSrc() PkgBuild {
	return src.Curl(d,
		Name("inetutils-src"),
		src.URL("http://ftp.gnu.org/gnu/inetutils/inetutils-1.9.4.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) IntltoolSrc() PkgBuild {
	return src.Curl(d,
		Name("intltool-src"),
		// TODO support mirrors, the original source used here is doen
		// src.URL("https://launchpad.net/intltool/trunk/0.51.0/+download/intltool-0.51.0.tar.gz"),
		src.URL("http://ftp.lfs-matrix.net/pub/clfs/conglomeration/intltool/intltool-0.51.0.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) IPRoute2Src() PkgBuild {
	return src.Curl(d,
		Name("iproute2-src"),
		src.URL("https://www.kernel.org/pub/linux/utils/net/iproute2/iproute2-5.2.0.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) KbdSrc() PkgBuild {
	return src.Curl(d,
		Name("kbd-src"),
		src.URL("https://www.kernel.org/pub/linux/utils/kbd/kbd-2.2.0.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LessSrc() PkgBuild {
	return src.Curl(d,
		Name("less-src"),
		src.URL("http://www.greenwoodsoftware.com/less/less-551.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LibcapSrc() PkgBuild {
	return src.Curl(d,
		Name("libcap-src"),
		src.URL("https://www.kernel.org/pub/linux/libs/security/linux-privs/libcap2/libcap-2.27.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LibffiSrc() PkgBuild {
	return src.Curl(d,
		Name("libffi-src"),
		src.URL("ftp://sourceware.org/pub/libffi/libffi-3.2.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LibpipelineSrc() PkgBuild {
	return src.Curl(d,
		Name("libpipeline-src"),
		src.URL("http://download.savannah.gnu.org/releases/libpipeline/libpipeline-1.5.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) LibtoolSrc() PkgBuild {
	return src.Curl(d,
		Name("libtool-src"),
		src.URL("http://ftp.gnu.org/gnu/libtool/libtool-2.4.6.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) ManDBSrc() PkgBuild {
	return src.Curl(d,
		Name("mandb-src"),
		src.URL("http://download.savannah.gnu.org/releases/man-db/man-db-2.8.6.1.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) ManpagesSrc() PkgBuild {
	return src.Curl(d,
		Name("manpages-src"),
		src.URL("https://www.kernel.org/pub/linux/docs/man-pages/man-pages-5.02.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) MesonSrc() PkgBuild {
	return src.Curl(d,
		Name("meson-src"),
		src.URL("https://github.com/mesonbuild/meson/releases/download/0.51.1/meson-0.51.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) NinjaSrc() PkgBuild {
	return src.Curl(d,
		Name("ninja-src"),
		src.URL("https://github.com/ninja-build/ninja/archive/v1.9.0/ninja-1.9.0.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) OpenSSLSrc() PkgBuild {
	return src.Curl(d,
		Name("openssl-src"),
		src.URL("https://www.openssl.org/source/openssl-1.1.1c.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) PkgConfigSrc() PkgBuild {
	return src.Curl(d,
		Name("pkgconfig-src"),
		src.URL("https://pkg-config.freedesktop.org/releases/pkg-config-0.29.2.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) Perl5XMLParserSrc() PkgBuild {
	return src.Curl(d,
		Name("perl5-xmlparser-src"),
		src.URL("https://cpan.metacpan.org/authors/id/T/TO/TODDR/XML-Parser-2.44.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) ProcpsSrc() PkgBuild {
	return src.Curl(d,
		Name("procps-src"),
		src.URL("https://sourceforge.net/projects/procps-ng/files/Production/procps-ng-3.3.15.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) PsmiscSrc() PkgBuild {
	return src.Curl(d,
		Name("psmisc-src"),
		src.URL("https://sourceforge.net/projects/psmisc/files/psmisc/psmisc-23.2.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) ReadlineSrc() PkgBuild {
	return src.Curl(d,
		Name("readline-src"),
		src.URL("http://ftp.gnu.org/gnu/readline/readline-8.0.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) TimezoneDataSrc() PkgBuild {
	return src.Curl(d,
		Name("timezonedata-src"),
		src.URL("https://www.iana.org/time-zones/repository/releases/tzdata2019b.tar.gz"),
		src.StripComponents(0),
	)
}

func (d distroSources) UtilLinuxSrc() PkgBuild {
	return src.Curl(d,
		Name("utillinux-src"),
		src.URL("https://www.kernel.org/pub/linux/utils/util-linux/v2.34/util-linux-2.34.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) ZlibSrc() PkgBuild {
	return src.Curl(d,
		Name("zlib-src"),
		src.URL("https://zlib.net/zlib-1.2.11.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) Libtasn1Src() PkgBuild {
	return src.Curl(d,
		Name("libtasn1-src"),
		src.URL("https://ftp.gnu.org/gnu/libtasn1/libtasn1-4.14.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) P11kitSrc() PkgBuild {
	return src.Curl(d,
		Name("p11kit-src"),
		src.URL("https://github.com/p11-glue/p11-kit/releases/download/0.23.16.1/p11-kit-0.23.16.1.tar.gz"),
		src.StripComponents(1),
	)
}

func (d distroSources) CACertsSrc() PkgBuild {
	return src.Curl(d,
		Name("cacerts-src"),
		src.URL("https://github.com/djlucas/make-ca/releases/download/v1.4/make-ca-1.4.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) CurlSrc() PkgBuild {
	return src.Curl(d,
		Name("curl-src"),
		src.URL("https://curl.haxx.se/download/curl-7.65.3.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GitSrc() PkgBuild {
	return src.Curl(d,
		Name("git-src"),
		src.URL("https://www.kernel.org/pub/software/scm/git/git-2.23.0.tar.xz"),
		src.StripComponents(1),
	)
}

func (d distroSources) GolangBootstrapSrc() PkgBuild {
	return src.Git(d,
		src.URL("https://github.com/golang/go.git"),
		src.Ref("release-branch.go1.4"),
		Name("golang-bootstrap-src"),
	)
}

func (d distroSources) GolangSrc() PkgBuild {
	return src.Git(d,
		src.URL("https://github.com/golang/go.git"),
		src.Ref("release-branch.go1.14"),
		Name("golang-src"),
		Deps(golang.BootstrapSrcPkg(d)),
	)
}

/*
func (d distroSources) HomedirSrc() homedir.SrcPkg {
	return homedir.SrcPkgOnce(d, func() Pkg {
return src.Git(d,
		Name("homedir-src"),
		src.URL("git@github.com:sipsma/home.git"),
		src.Ref("rootless"),
	))

}
*/
