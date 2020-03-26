package distro

import (
	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"

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
	"github.com/sipsma/bincastle/distro/pkgs/utillinux"
	"github.com/sipsma/bincastle/distro/pkgs/timezonedata"
	"github.com/sipsma/bincastle/distro/pkgs/xz"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	"github.com/sipsma/bincastle/distro/pkgs/p11kit"
	"github.com/sipsma/bincastle/distro/pkgs/nettle"
	"github.com/sipsma/bincastle/distro/pkgs/libunistring"
	"github.com/sipsma/bincastle/distro/pkgs/gnutls"
)

type distroSources struct {
	Pkger
}

func (d distroSources) AwkSrc() awk.SrcPkg {
	return awk.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"awk-src",
			"https://ftp.gnu.org/gnu/gawk/gawk-5.0.1.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) BashSrc() bash.SrcPkg {
	return bash.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"bash-src",
			"https://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) BinutilsSrc() binutils.SrcPkg {
	return binutils.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"binutils-src",
			"https://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) BisonSrc() bison.SrcPkg {
	return bison.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"bison-src",
			"https://ftp.gnu.org/gnu/bison/bison-3.4.1.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) Bzip2Src() bzip2.SrcPkg {
	return bzip2.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"bzip2-src",
			"https://www.sourceware.org/pub/bzip2/bzip2-1.0.8.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) CoreutilsSrc() coreutils.SrcPkg {
	return coreutils.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"coreutils-src",
			"https://ftp.gnu.org/gnu/coreutils/coreutils-8.31.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) DiffutilsSrc() diffutils.SrcPkg {
	return diffutils.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"diffutils-src",
			"https://ftp.gnu.org/gnu/diffutils/diffutils-3.7.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) FileSrc() file.SrcPkg {
	return file.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"file-src",
			"ftp://ftp.astron.com/pub/file/file-5.37.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) FindutilsSrc() findutils.SrcPkg {
	return findutils.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"findutils-src",
			"https://ftp.gnu.org/gnu/findutils/findutils-4.6.0.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GCCSrc() gcc.SrcPkg {
	return gcc.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gcc-src",
			"http://ftp.gnu.org/gnu/gcc/gcc-9.2.0/gcc-9.2.0.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GettextSrc() gettext.SrcPkg {
	return gettext.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gettext-src",
			"https://ftp.gnu.org/gnu/gettext/gettext-0.20.1.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GMPSrc() gmp.SrcPkg {
	return gmp.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gmp-src",
			"https://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GrepSrc() grep.SrcPkg {
	return grep.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"grep-src",
			"https://ftp.gnu.org/gnu/grep/grep-3.3.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GzipSrc() gzip.SrcPkg {
	return gzip.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gzip-src",
			"https://ftp.gnu.org/gnu/gzip/gzip-1.10.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibcSrc() libc.SrcPkg {
	return libc.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libc-src",
			"http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibstdcppSrc() libstdcpp.SrcPkg {
	return libstdcpp.BuildSrcPkg(d, func() Pkg {
		// TODO you should change the outputdir to be down a
		// level at ./libstdc++. This will require changing
		// OutputDir to append if provided a relpath.
		return d.GCCSrc()
	})
}

func (d distroSources) LinuxSrc() linux.SrcPkg {
	return linux.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"linux-src",
			"https://www.kernel.org/pub/linux/kernel/v5.x/linux-5.2.8.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) M4Src() m4.SrcPkg {
	return m4.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"m4-src",
			"https://ftp.gnu.org/gnu/m4/m4-1.4.18.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) MakeSrc() make.SrcPkg {
	return make.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"make-src",
			"https://ftp.gnu.org/gnu/make/make-4.2.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) MPCSrc() mpc.SrcPkg {
	return mpc.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"mpc-src",
			"https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) MPFRSrc() mpfr.SrcPkg {
	return mpfr.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"mpfr-src",
			"https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) NcursesSrc() ncurses.SrcPkg {
	return ncurses.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"ncurses-src",
			"https://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) PatchSrc() patch.SrcPkg {
	return patch.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"patch-src",
			"https://ftp.gnu.org/gnu/patch/patch-2.7.6.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) Perl5Src() perl5.SrcPkg {
	return perl5.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"perl5-src",
			"https://www.cpan.org/src/5.0/perl-5.30.0.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) Python3Src() python3.SrcPkg {
	return python3.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"python3-src",
			"https://www.python.org/ftp/python/3.7.4/Python-3.7.4.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) SedSrc() sed.SrcPkg {
	return sed.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"sed-src",
			"https://ftp.gnu.org/gnu/sed/sed-4.7.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) TarSrc() tar.SrcPkg {
	return tar.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"tar-src",
			"https://ftp.gnu.org/gnu/tar/tar-1.32.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) TexinfoSrc() texinfo.SrcPkg {
	return texinfo.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"texinfo-src",
			"https://ftp.gnu.org/gnu/texinfo/texinfo-6.6.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) XzSrc() xz.SrcPkg {
	return xz.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"xz-src",
			"https://astuteinternet.dl.sourceforge.net/project/lzmautils/xz-5.2.4.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) AclSrc() acl.SrcPkg {
	return acl.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"acl-src",
			"http://download.savannah.gnu.org/releases/acl/acl-2.2.53.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) AttrSrc() attr.SrcPkg {
	return attr.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"attr-src",
			"http://download.savannah.gnu.org/releases/attr/attr-2.4.48.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) AutoconfSrc() autoconf.SrcPkg {
	return autoconf.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"autoconf-src",
			"http://ftp.gnu.org/gnu/autoconf/autoconf-2.69.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) AutomakeSrc() automake.SrcPkg {
	return automake.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"automake-src",
			"http://ftp.gnu.org/gnu/automake/automake-1.16.1.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) BCSrc() bc.SrcPkg {
	return bc.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"bc-src",
			"https://github.com/gavinhoward/bc/archive/2.1.3/bc-2.1.3.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) E2fsprogsSrc() e2fsprogs.SrcPkg {
	return e2fsprogs.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"e2fsprogs-src",
			"https://downloads.sourceforge.net/project/e2fsprogs/e2fsprogs/v1.45.3/e2fsprogs-1.45.3.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ElfutilsSrc() elfutils.SrcPkg {
	return elfutils.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"elfutils-src",
			"https://sourceware.org/ftp/elfutils/0.177/elfutils-0.177.tar.bz2",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ExpatSrc() expat.SrcPkg {
	return expat.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"expat-src",
			"https://prdownloads.sourceforge.net/expat/expat-2.2.7.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) FlexSrc() flex.SrcPkg {
	return flex.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"flex-src",
			"https://github.com/westes/flex/releases/download/v2.6.4/flex-2.6.4.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GDBMSrc() gdbm.SrcPkg {
	return gdbm.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gdbm-src",
			"http://ftp.gnu.org/gnu/gdbm/gdbm-1.18.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GperfSrc() gperf.SrcPkg {
	return gperf.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gperf-src",
			"http://ftp.gnu.org/gnu/gperf/gperf-3.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GroffSrc() groff.SrcPkg {
	return groff.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"groff-src",
			"http://ftp.gnu.org/gnu/groff/groff-1.22.4.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) IanaetcSrc() ianaetc.SrcPkg {
	return ianaetc.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"ianaetc-src",
			"http://anduin.linuxfromscratch.org/LFS/iana-etc-2.30.tar.bz2",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) InetutilsSrc() inetutils.SrcPkg {
	return inetutils.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"inetutils-src",
			"http://ftp.gnu.org/gnu/inetutils/inetutils-1.9.4.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) IntltoolSrc() intltool.SrcPkg {
	return intltool.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"intltool-src",
			// TODO support mirrors, the original source used here is doen
			// src.URL("https://launchpad.net/intltool/trunk/0.51.0/+download/intltool-0.51.0.tar.gz"),
			"http://ftp.lfs-matrix.net/pub/clfs/conglomeration/intltool/intltool-0.51.0.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) IPRoute2Src() iproute2.SrcPkg {
	return iproute2.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"iproute2-src",
			"https://www.kernel.org/pub/linux/utils/net/iproute2/iproute2-5.2.0.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) KbdSrc() kbd.SrcPkg {
	return kbd.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"kbd-src",
			"https://www.kernel.org/pub/linux/utils/kbd/kbd-2.2.0.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LessSrc() less.SrcPkg {
	return less.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"less-src",
			"http://www.greenwoodsoftware.com/less/less-551.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibcapSrc() libcap.SrcPkg {
	return libcap.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libcap-src",
			"https://www.kernel.org/pub/linux/libs/security/linux-privs/libcap2/libcap-2.27.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibffiSrc() libffi.SrcPkg {
	return libffi.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libffi-src",
			"ftp://sourceware.org/pub/libffi/libffi-3.2.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibpipelineSrc() libpipeline.SrcPkg {
	return libpipeline.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libpipeline-src",
			"http://download.savannah.gnu.org/releases/libpipeline/libpipeline-1.5.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibtoolSrc() libtool.SrcPkg {
	return libtool.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libtool-src",
			"http://ftp.gnu.org/gnu/libtool/libtool-2.4.6.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ManDBSrc() mandb.SrcPkg {
	return mandb.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"mandb-src",
			"http://download.savannah.gnu.org/releases/man-db/man-db-2.8.6.1.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ManpagesSrc() manpages.SrcPkg {
	return manpages.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"manpages-src",
			"https://www.kernel.org/pub/linux/docs/man-pages/man-pages-5.02.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) MesonSrc() meson.SrcPkg {
	return meson.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"meson-src",
			"https://github.com/mesonbuild/meson/releases/download/0.51.1/meson-0.51.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) NinjaSrc() ninja.SrcPkg {
	return ninja.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"ninja-src",
			"https://github.com/ninja-build/ninja/archive/v1.9.0/ninja-1.9.0.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) OpenSSLSrc() openssl.SrcPkg {
	return openssl.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"openssl-src",
			"https://www.openssl.org/source/openssl-1.1.1c.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) PkgConfigSrc() pkgconfig.SrcPkg {
	return pkgconfig.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"pkgconfig-src",
			"https://pkg-config.freedesktop.org/releases/pkg-config-0.29.2.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) Perl5XMLParserSrc() perl5.XMLParserSrcPkg {
	return perl5.BuildXMLParserSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"perl5-xmlparser-src",
			"https://cpan.metacpan.org/authors/id/T/TO/TODDR/XML-Parser-2.44.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ProcpsSrc() procps.SrcPkg {
	return procps.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"procps-src",
			"https://sourceforge.net/projects/procps-ng/files/Production/procps-ng-3.3.15.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) PsmiscSrc() psmisc.SrcPkg {
	return psmisc.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"psmisc-src",
			"https://sourceforge.net/projects/psmisc/files/psmisc/psmisc-23.2.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ReadlineSrc() readline.SrcPkg {
	return readline.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"readline-src",
			"http://ftp.gnu.org/gnu/readline/readline-8.0.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) TimezoneDataSrc() timezonedata.SrcPkg {
	return timezonedata.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"timezonedata-src",
			"https://www.iana.org/time-zones/repository/releases/tzdata2019b.tar.gz",
			src.CurlOpt{
				StripComponents: 0,
			},
		)
	})
}

func (d distroSources) UtilLinuxSrc() utillinux.SrcPkg {
	return utillinux.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"utillinux-src",
			"https://www.kernel.org/pub/linux/utils/util-linux/v2.34/util-linux-2.34.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) ZlibSrc() zlib.SrcPkg {
	return zlib.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"zlib-src",
			"https://zlib.net/zlib-1.2.11.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) Libtasn1Src() libtasn1.SrcPkg {
	return libtasn1.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libtasn1-src",
			"https://ftp.gnu.org/gnu/libtasn1/libtasn1-4.14.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) P11kitSrc() p11kit.SrcPkg {
	return p11kit.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"p11kit-src",
			"https://github.com/p11-glue/p11-kit/releases/download/0.23.16.1/p11-kit-0.23.16.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) CACertsSrc() cacerts.SrcPkg {
	return cacerts.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"cacerts-src",
			"https://github.com/djlucas/make-ca/releases/download/v1.4/make-ca-1.4.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) CurlSrc() curl.SrcPkg {
	return curl.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"curl-src",
			"https://curl.haxx.se/download/curl-7.65.3.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GitSrc() git.SrcPkg {
	return git.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"git-src",
			"https://www.kernel.org/pub/software/scm/git/git-2.23.0.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GolangBootstrapSrc() golang.BootstrapSrcPkg {
	return golang.BuildBootstrapSrcPkg(d, func() Pkg {
		return src.Git(d,
			"golang-bootstrap-src",
			"https://github.com/golang/go.git",
			src.GitOpt{
				Ref: "release-branch.go1.4",
			},
		)
	})
}

func (d distroSources) GolangSrc() golang.SrcPkg {
	return golang.BuildSrcPkg(d, func() Pkg {
		return src.Git(d,
			"golang-src",
			"https://github.com/golang/go.git",
			src.GitOpt{
				Ref: "release-branch.go1.14",
			},
		).With(RuntimeDeps(d.GolangBootstrapSrc()))
	})
}

func (d distroSources) NettleSrc() nettle.SrcPkg {
	return nettle.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"nettle-src",
			"https://ftp.gnu.org/gnu/nettle/nettle-3.5.1.tar.gz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibunistringSrc() libunistring.SrcPkg {
	return libunistring.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"libunistring-src",
			"https://ftp.gnu.org/gnu/libunistring/libunistring-0.9.10.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) GNUTLSSrc() gnutls.SrcPkg {
	return gnutls.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"gnutls-src",
			"https://www.gnupg.org/ftp/gcrypt/gnutls/v3.6/gnutls-3.6.12.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) EmacsSrc() emacs.SrcPkg {
	return emacs.BuildSrcPkg(d, func() Pkg {
		return src.Curl(d,
			"emacs-src",
			"https://ftp.gnu.org/gnu/emacs/emacs-26.3.tar.xz",
			src.CurlOpt{
				StripComponents: 1,
			},
		)
	})
}

func (d distroSources) LibeventSrc() Pkg {
	return src.Git(d,
		"libevent-src",
		"https://github.com/libevent/libevent.git",
		src.GitOpt{
			Ref: "release-2.1.11-stable",
		},
	)
}

func (d distroSources) TmuxSrc() Pkg {
	return src.Git(d,
		"tmux-src",
		"https://github.com/tmux/tmux.git",
		src.GitOpt{
			Ref: "3.0",
		},
	)
}

func (d distroSources) WhichSrc() Pkg {
	return src.Curl(d,
		"which-src",
		"https://ftp.gnu.org/gnu/which/which-2.21.tar.gz",
		src.CurlOpt{
			StripComponents: 1,
		},
	)
}

func (d distroSources) OpenSSHSrc() Pkg {
	return src.Curl(d,
		"openssh-src",
		"http://ftp.openbsd.org/pub/OpenBSD/OpenSSH/portable/openssh-8.2p1.tar.gz",
		src.CurlOpt{
			StripComponents: 1,
		},
	)
}
