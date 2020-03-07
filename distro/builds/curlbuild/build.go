package curlbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/autoconf"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/cacerts"
	"github.com/sipsma/bincastle/distro/pkgs/curl"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	"github.com/sipsma/bincastle/distro/pkgs/openssl"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	curl.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	m4.Pkger
	autoconf.Pkger
	openssl.Pkger
	zlib.Pkger
	cacerts.Pkger
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		linux.HeadersPkg(d),
		libc.Pkg(d),
		binutils.Pkg(d),
		gcc.Pkg(d),
		pkgconfig.Pkg(d),
		m4.Pkg(d),
		autoconf.Pkg(d),
		openssl.Pkg(d),
		zlib.Pkg(d),
		cacerts.Pkg(d),
		curl.SrcPkg(d),
		ScratchMount(`/build`),
		Shell(
			`cd /build`,
			strings.Join([]string{
				`/src/curl-src/configure`,
				`--prefix=/usr`,
				`--disable-static`,
				`--enable-threaded-resolver`,
				`--with-ca-path=/etc/ssl/certs`,
			}, " "),
			`make`,
			`make install`,
		),
	).With(
		Name("curl"),
		Deps(
			libc.Pkg(d),
			openssl.Pkg(d),
			zlib.Pkg(d),
			cacerts.Pkg(d),
		),
	).With(opts...))
}
