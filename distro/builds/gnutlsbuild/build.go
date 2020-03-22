package gnutlsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libunistring"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
	"github.com/sipsma/bincastle/distro/pkgs/cacerts"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libffi"
	"github.com/sipsma/bincastle/distro/pkgs/libtasn1"
	"github.com/sipsma/bincastle/distro/pkgs/p11kit"
	"github.com/sipsma/bincastle/distro/pkgs/nettle"
	"github.com/sipsma/bincastle/distro/pkgs/gnutls"
)

func Default(d interface {
	PkgCache
	Executor
	gnutls.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	nettle.Pkger
	libunistring.Pkger
	cacerts.Pkger
	libtasn1.Pkger
	p11kit.Pkger
	gmp.Pkger
	libffi.Pkger
}, opts ...Opt) gnutls.Pkg {
	return gnutls.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.Nettle(),
				d.Libunistring(),
				d.CACerts(),
				d.Libtasn1(),
				d.P11kit(),
				d.GMP(),
				d.Libffi(),
				d.GNUTLSSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/gnutls-src/configure`,
					`--prefix=/usr`,
					`--docdir=/usr/share/doc/gnutls-3.6.12`,
					`--disable-guile`,
					`--with-default-trust-store-pkcs11="pkcs11:"`,
				}, " "),
				`make`,
				`make install`,
			),
		).With(
			Name("gnutls"),
			RuntimeDeps(
				d.Libc(),
				d.GCC(),
				d.Nettle(),
				d.Libunistring(),
				d.Libtasn1(),
				d.P11kit(),
				d.GMP(),
				d.Libffi(),
				d.CACerts(),
			),
		).With(opts...)
	})
}
