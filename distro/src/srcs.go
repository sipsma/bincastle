package src

import (
	"fmt"
	"path/filepath"

	"github.com/sipsma/bincastle/distro/bootstrap"
	. "github.com/sipsma/bincastle/graph"
)

type ViaCurl struct {
	URL             string
	Name            string
	StripComponents int
	AlwaysRun       AlwaysRun
}

func (s ViaCurl) Spec() Spec {
	return LayerSpec(
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		Shell(
			`mkdir -p /src`,
			`cd /src`,
			fmt.Sprintf("curl -L -O %s", s.URL),
			`DLFILE=$(ls)`,
			fmt.Sprintf(
				`tar --strip-components=%d --extract --no-same-owner --file=$DLFILE`,
				s.StripComponents),
			`rm $DLFILE`,
		),
		s.AlwaysRun,
		OutputDir("/src"),
		MountDir(filepath.Join("/src", s.Name)),
	)
}

type ViaGit struct {
	URL       string
	Ref       string
	Name      string
	AlwaysRun AlwaysRun
	WithSSH   bool
}

func (s ViaGit) Spec() Spec {
	var extras []LayerSpecOpt
	if s.WithSSH {
		extras = append(extras, Env("SSH_AUTH_SOCK", "/run/ssh-agent.sock"))
	}

	return LayerSpec(
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		Shell(
			`mkdir -p /src`,
			fmt.Sprintf(`git clone --recurse-submodules %s /src`, s.URL),
			`cd /src`,
			fmt.Sprintf(`git checkout %s`, s.Ref),
		),
		s.AlwaysRun,
		MergeLayerSpecOpts(extras...),
		OutputDir("/src"),
		MountDir(filepath.Join("/src", s.Name)),
	)
}

type Awk struct{}

func (Awk) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/gawk/gawk-5.0.1.tar.xz",
		Name:            "awk-src",
		StripComponents: 1,
	}.Spec()
}

type Bash struct{}

func (Bash) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz",
		Name:            "bash-src",
		StripComponents: 1,
	}.Spec()
}

type Binutils struct{}

func (Binutils) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz",
		Name:            "binutils-src",
		StripComponents: 1,
	}.Spec()
}

type Bison struct{}

func (Bison) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/bison/bison-3.4.1.tar.xz",
		Name:            "bison-src",
		StripComponents: 1,
	}.Spec()
}

type Bzip2 struct{}

func (Bzip2) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.sourceware.org/pub/bzip2/bzip2-1.0.8.tar.gz",
		Name:            "bzip2-src",
		StripComponents: 1,
	}.Spec()
}

type Coreutils struct{}

func (Coreutils) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/coreutils/coreutils-8.31.tar.xz",
		Name:            "coreutils-src",
		StripComponents: 1,
	}.Spec()
}

type Diffutils struct{}

func (Diffutils) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/diffutils/diffutils-3.7.tar.xz",
		Name:            "diffutils-src",
		StripComponents: 1,
	}.Spec()
}

type File struct{}

func (File) Spec() Spec {
	return ViaCurl{
		URL:             "ftp://ftp.astron.com/pub/file/file-5.37.tar.gz",
		Name:            "file-src",
		StripComponents: 1,
	}.Spec()
}

type Findutils struct{}

func (Findutils) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/findutils/findutils-4.6.0.tar.gz",
		Name:            "findutils-src",
		StripComponents: 1,
	}.Spec()
}

type GCC struct{}

func (GCC) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/gcc/gcc-9.2.0/gcc-9.2.0.tar.xz",
		Name:            "gcc-src",
		StripComponents: 1,
	}.Spec()
}

type Gettext struct{}

func (Gettext) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/gettext/gettext-0.20.1.tar.xz",
		Name:            "gettext-src",
		StripComponents: 1,
	}.Spec()
}

type GMP struct{}

func (GMP) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/gmp/gmp-6.1.2.tar.xz",
		Name:            "gmp-src",
		StripComponents: 1,
	}.Spec()
}

type Grep struct{}

func (Grep) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/grep/grep-3.3.tar.xz",
		Name:            "grep-src",
		StripComponents: 1,
	}.Spec()
}

type Gzip struct{}

func (Gzip) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/gzip/gzip-1.10.tar.xz",
		Name:            "gzip-src",
		StripComponents: 1,
	}.Spec()
}

type Libc struct{}

func (Libc) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz",
		Name:            "libc-src",
		StripComponents: 1,
	}.Spec()
}

type Linux struct{}

func (Linux) Spec() Spec {
	return LayerSpec(
		Dep(ViaCurl{
			URL:             "https://www.kernel.org/pub/linux/kernel/v5.x/linux-5.2.8.tar.xz",
			Name:            "linux-src",
			StripComponents: 1,
		}),
		BuildDep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		Shell(
			`cd /src/linux-src`,
			`make mrproper`,
		),
	)
}

type M4 struct{}

func (M4) Spec() Spec {
	return LayerSpec(
		Dep(ViaCurl{
			URL:             "https://ftp.gnu.org/gnu/m4/m4-1.4.18.tar.xz",
			Name:            "m4-src",
			StripComponents: 1,
		}),
		BuildDep(bootstrap.Spec{}),
		Shell(
			`cd /src/m4-src`,
			`sed -i 's/IO_ftrylockfile/IO_EOF_SEEN/' lib/*.c`,
			`echo "#define _IO_IN_BACKUP 0x100" >> lib/stdio-impl.h`,
		),
	)
}

type Make struct{}

func (Make) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/make/make-4.2.1.tar.gz",
		Name:            "make-src",
		StripComponents: 1,
	}.Spec()
}

type MPC struct{}

func (MPC) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/mpc/mpc-1.1.0.tar.gz",
		Name:            "mpc-src",
		StripComponents: 1,
	}.Spec()
}

type MPFR struct{}

func (MPFR) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.mpfr.org/mpfr-4.0.2/mpfr-4.0.2.tar.xz",
		Name:            "mpfr-src",
		StripComponents: 1,
	}.Spec()
}

type Ncurses struct{}

func (Ncurses) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz",
		Name:            "ncurses-src",
		StripComponents: 1,
	}.Spec()
}

type Patch struct{}

func (Patch) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/patch/patch-2.7.6.tar.xz",
		Name:            "patch-src",
		StripComponents: 1,
	}.Spec()
}

type Perl5 struct{}

func (Perl5) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.cpan.org/src/5.0/perl-5.30.0.tar.xz",
		Name:            "perl5-src",
		StripComponents: 1,
	}.Spec()
}

type Python3 struct{}

func (Python3) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.python.org/ftp/python/3.7.4/Python-3.7.4.tar.xz",
		Name:            "python3-src",
		StripComponents: 1,
	}.Spec()
}

type Sed struct{}

func (Sed) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/sed/sed-4.7.tar.xz",
		Name:            "sed-src",
		StripComponents: 1,
	}.Spec()
}

type Tar struct{}

func (Tar) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/tar/tar-1.32.tar.xz",
		Name:            "tar-src",
		StripComponents: 1,
	}.Spec()
}

type Texinfo struct{}

func (Texinfo) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/texinfo/texinfo-6.6.tar.xz",
		Name:            "texinfo-src",
		StripComponents: 1,
	}.Spec()
}

type Xz struct{}

func (Xz) Spec() Spec {
	return ViaCurl{
		URL:             "https://astuteinternet.dl.sourceforge.net/project/lzmautils/xz-5.2.4.tar.xz",
		Name:            "xz-src",
		StripComponents: 1,
	}.Spec()
}

type Acl struct{}

func (Acl) Spec() Spec {
	return ViaCurl{
		URL:             "http://download.savannah.gnu.org/releases/acl/acl-2.2.53.tar.gz",
		Name:            "acl-src",
		StripComponents: 1,
	}.Spec()
}

type Attr struct{}

func (Attr) Spec() Spec {
	return ViaCurl{
		URL:             "http://download.savannah.gnu.org/releases/attr/attr-2.4.48.tar.gz",
		Name:            "attr-src",
		StripComponents: 1,
	}.Spec()
}

type Autoconf struct{}

func (Autoconf) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/autoconf/autoconf-2.69.tar.xz",
		Name:            "autoconf-src",
		StripComponents: 1,
	}.Spec()
}

type Automake struct{}

func (Automake) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/automake/automake-1.16.1.tar.xz",
		Name:            "automake-src",
		StripComponents: 1,
	}.Spec()
}

type BC struct{}

func (BC) Spec() Spec {
	return ViaCurl{
		URL:             "https://github.com/gavinhoward/bc/archive/2.1.3/bc-2.1.3.tar.gz",
		Name:            "bc-src",
		StripComponents: 1,
	}.Spec()
}

type E2fsprogs struct{}

func (E2fsprogs) Spec() Spec {
	return ViaCurl{
		URL:             "https://downloads.sourceforge.net/project/e2fsprogs/e2fsprogs/v1.45.3/e2fsprogs-1.45.3.tar.gz",
		Name:            "e2fsprogs-src",
		StripComponents: 1,
	}.Spec()
}

type Elfutils struct{}

func (Elfutils) Spec() Spec {
	return ViaCurl{
		URL:             "https://sourceware.org/ftp/elfutils/0.177/elfutils-0.177.tar.bz2",
		Name:            "elfutils-src",
		StripComponents: 1,
	}.Spec()
}

type Expat struct{}

func (Expat) Spec() Spec {
	return ViaCurl{
		URL:             "https://prdownloads.sourceforge.net/expat/expat-2.2.7.tar.xz",
		Name:            "expat-src",
		StripComponents: 1,
	}.Spec()
}

type Flex struct{}

func (Flex) Spec() Spec {
	return ViaCurl{
		URL:             "https://github.com/westes/flex/releases/download/v2.6.4/flex-2.6.4.tar.gz",
		Name:            "flex-src",
		StripComponents: 1,
	}.Spec()
}

type GDBM struct{}

func (GDBM) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/gdbm/gdbm-1.18.1.tar.gz",
		Name:            "gdbm-src",
		StripComponents: 1,
	}.Spec()
}

type Gperf struct{}

func (Gperf) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/gperf/gperf-3.1.tar.gz",
		Name:            "gperf-src",
		StripComponents: 1,
	}.Spec()
}

type Groff struct{}

func (Groff) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/groff/groff-1.22.4.tar.gz",
		Name:            "groff-src",
		StripComponents: 1,
	}.Spec()
}

type Ianaetc struct{}

func (Ianaetc) Spec() Spec {
	return ViaCurl{
		URL:             "http://anduin.linuxfromscratch.org/LFS/iana-etc-2.30.tar.bz2",
		Name:            "ianaetc-src",
		StripComponents: 1,
	}.Spec()
}

type Inetutils struct{}

func (Inetutils) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/inetutils/inetutils-1.9.4.tar.xz",
		Name:            "inetutils-src",
		StripComponents: 1,
	}.Spec()
}

type Intltool struct{}

func (Intltool) Spec() Spec {
	return ViaCurl{
		// TODO support mirrors, the original source used here is down
		// "https://launchpad.net/intltool/trunk/0.51.0/+download/intltool-0.51.0.tar.gz"
		URL:             "http://ftp.lfs-matrix.net/pub/clfs/conglomeration/intltool/intltool-0.51.0.tar.gz",
		Name:            "intltool-src",
		StripComponents: 1,
	}.Spec()
}

type IPRoute2 struct{}

func (IPRoute2) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.kernel.org/pub/linux/utils/net/iproute2/iproute2-5.2.0.tar.xz",
		Name:            "iproute2-src",
		StripComponents: 1,
	}.Spec()
}

type Kbd struct{}

func (Kbd) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.kernel.org/pub/linux/utils/kbd/kbd-2.2.0.tar.xz",
		Name:            "kbd-src",
		StripComponents: 1,
	}.Spec()
}

type Less struct{}

func (Less) Spec() Spec {
	return ViaCurl{
		URL:             "http://www.greenwoodsoftware.com/less/less-551.tar.gz",
		Name:            "less-src",
		StripComponents: 1,
	}.Spec()
}

type Libcap struct{}

func (Libcap) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.kernel.org/pub/linux/libs/security/linux-privs/libcap2/libcap-2.27.tar.xz",
		Name:            "libcap-src",
		StripComponents: 1,
	}.Spec()
}

type Libffi struct{}

func (Libffi) Spec() Spec {
	return ViaCurl{
		URL:             "ftp://sourceware.org/pub/libffi/libffi-3.2.1.tar.gz",
		Name:            "libffi-src",
		StripComponents: 1,
	}.Spec()
}

type Libpipeline struct{}

func (Libpipeline) Spec() Spec {
	return ViaCurl{
		URL:             "http://download.savannah.gnu.org/releases/libpipeline/libpipeline-1.5.1.tar.gz",
		Name:            "libpipeline-src",
		StripComponents: 1,
	}.Spec()
}

type Libtool struct{}

func (Libtool) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/libtool/libtool-2.4.6.tar.xz",
		Name:            "libtool-src",
		StripComponents: 1,
	}.Spec()
}

type ManDB struct{}

func (ManDB) Spec() Spec {
	return ViaCurl{
		URL:             "http://download.savannah.gnu.org/releases/man-db/man-db-2.8.6.1.tar.xz",
		Name:            "mandb-src",
		StripComponents: 1,
	}.Spec()
}

type Manpages struct{}

func (Manpages) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.kernel.org/pub/linux/docs/man-pages/man-pages-5.02.tar.xz",
		Name:            "manpages-src",
		StripComponents: 1,
	}.Spec()
}

type Meson struct{}

func (Meson) Spec() Spec {
	return ViaCurl{
		URL:             "https://github.com/mesonbuild/meson/releases/download/0.51.1/meson-0.51.1.tar.gz",
		Name:            "meson-src",
		StripComponents: 1,
	}.Spec()
}

type Ninja struct{}

func (Ninja) Spec() Spec {
	return ViaCurl{
		URL:             "https://github.com/ninja-build/ninja/archive/v1.9.0/ninja-1.9.0.tar.gz",
		Name:            "ninja-src",
		StripComponents: 1,
	}.Spec()
}

type OpenSSL struct{}

func (OpenSSL) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.openssl.org/source/openssl-1.1.1f.tar.gz",
		Name:            "openssl-src",
		StripComponents: 1,
	}.Spec()
}

type PkgConfig struct{}

func (PkgConfig) Spec() Spec {
	return ViaCurl{
		URL:             "https://pkg-config.freedesktop.org/releases/pkg-config-0.29.2.tar.gz",
		Name:            "pkgconfig-src",
		StripComponents: 1,
	}.Spec()
}

type Perl5XMLParser struct{}

func (Perl5XMLParser) Spec() Spec {
	return ViaCurl{
		URL:             "https://cpan.metacpan.org/authors/id/T/TO/TODDR/XML-Parser-2.44.tar.gz",
		Name:            "perl5-xmlparser-src",
		StripComponents: 1,
	}.Spec()
}

type Procps struct{}

func (Procps) Spec() Spec {
	return ViaCurl{
		URL:             "https://sourceforge.net/projects/procps-ng/files/Production/procps-ng-3.3.15.tar.xz",
		Name:            "procps-src",
		StripComponents: 1,
	}.Spec()
}

type Psmisc struct{}

func (Psmisc) Spec() Spec {
	return ViaCurl{
		URL:             "https://sourceforge.net/projects/psmisc/files/psmisc/psmisc-23.2.tar.xz",
		Name:            "psmisc-src",
		StripComponents: 1,
	}.Spec()
}

type Readline struct{}

func (Readline) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.gnu.org/gnu/readline/readline-8.0.tar.gz",
		Name:            "readline-src",
		StripComponents: 1,
	}.Spec()
}

type TimezoneData struct{}

func (TimezoneData) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.iana.org/time-zones/repository/releases/tzdata2019b.tar.gz",
		Name:            "timezonedata-src",
		StripComponents: 0,
	}.Spec()
}

type UtilLinux struct{}

func (UtilLinux) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.kernel.org/pub/linux/utils/util-linux/v2.34/util-linux-2.34.tar.xz",
		Name:            "utillinux-src",
		StripComponents: 1,
	}.Spec()
}

type Zlib struct{}

func (Zlib) Spec() Spec {
	return ViaCurl{
		URL:             "https://zlib.net/zlib-1.2.11.tar.xz",
		Name:            "zlib-src",
		StripComponents: 1,
	}.Spec()
}

type Libtasn1 struct{}

func (Libtasn1) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/libtasn1/libtasn1-4.14.tar.gz",
		Name:            "libtasn1-src",
		StripComponents: 1,
	}.Spec()
}

type P11kit struct{}

func (P11kit) Spec() Spec {
	return ViaCurl{
		URL:             "https://github.com/p11-glue/p11-kit/releases/download/0.23.16.1/p11-kit-0.23.16.1.tar.gz",
		Name:            "p11kit-src",
		StripComponents: 1,
	}.Spec()
}

type CACerts struct{}

func (CACerts) Spec() Spec {
	return ViaCurl{
		URL:             "https://github.com/djlucas/make-ca/releases/download/v1.4/make-ca-1.4.tar.xz",
		Name:            "cacerts-src",
		StripComponents: 1,
	}.Spec()
}

type Curl struct{}

func (Curl) Spec() Spec {
	return ViaCurl{
		URL:             "https://curl.haxx.se/download/curl-7.65.3.tar.xz",
		Name:            "curl-src",
		StripComponents: 1,
	}.Spec()
}

type Git struct{}

func (Git) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.kernel.org/pub/software/scm/git/git-2.23.0.tar.xz",
		Name:            "git-src",
		StripComponents: 1,
	}.Spec()
}

type Golang struct{}

func (Golang) Spec() Spec {
	return Merge(
		ViaGit{
			URL:  "https://github.com/golang/go.git",
			Ref:  "release-branch.go1.4",
			Name: "golang-bootstrap-src",
		},
		ViaGit{
			URL:  "https://github.com/golang/go.git",
			Ref:  "release-branch.go1.14",
			Name: "golang-src",
		},
	)
}

type Nettle struct{}

func (Nettle) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/nettle/nettle-3.5.1.tar.gz",
		Name:            "nettle-src",
		StripComponents: 1,
	}.Spec()
}

type Libunistring struct{}

func (Libunistring) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/libunistring/libunistring-0.9.10.tar.xz",
		Name:            "libunistring-src",
		StripComponents: 1,
	}.Spec()
}

type GNUTLS struct{}

func (GNUTLS) Spec() Spec {
	return ViaCurl{
		URL:             "https://www.gnupg.org/ftp/gcrypt/gnutls/v3.6/gnutls-3.6.12.tar.xz",
		Name:            "gnutls-src",
		StripComponents: 1,
	}.Spec()
}

type Emacs struct{}

func (Emacs) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/emacs/emacs-26.3.tar.xz",
		Name:            "emacs-src",
		StripComponents: 1,
	}.Spec()
}

type Libevent struct{}

func (Libevent) Spec() Spec {
	return ViaGit{
		URL:  "https://github.com/libevent/libevent.git",
		Ref:  "release-2.1.11-stable",
		Name: "libevent-src",
	}.Spec()
}

type Tmux struct{}

func (Tmux) Spec() Spec {
	return ViaGit{
		URL:  "https://github.com/tmux/tmux.git",
		Ref:  "3.0",
		Name: "tmux-src",
	}.Spec()
}

type Which struct{}

func (Which) Spec() Spec {
	return ViaCurl{
		URL:             "https://ftp.gnu.org/gnu/which/which-2.21.tar.gz",
		Name:            "which-src",
		StripComponents: 1,
	}.Spec()
}

type OpenSSH struct{}

func (OpenSSH) Spec() Spec {
	return ViaCurl{
		URL:             "http://ftp.openbsd.org/pub/OpenBSD/OpenSSH/portable/openssh-8.2p1.tar.gz",
		Name:            "openssh-src",
		StripComponents: 1,
	}.Spec()
}
