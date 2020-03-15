package p11kitbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/automake"
	"github.com/sipsma/bincastle/distro/pkgs/binutils"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libffi"
	"github.com/sipsma/bincastle/distro/pkgs/libtasn1"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/m4"
	"github.com/sipsma/bincastle/distro/pkgs/p11kit"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	p11kit.Srcer
	linux.HeadersPkger
	libc.Pkger
	binutils.Pkger
	gcc.Pkger
	pkgconfig.Pkger
	m4.Pkger
	automake.Pkger
	libffi.Pkger
	libtasn1.Pkger
}, opts ...Opt) p11kit.Pkg {
	return p11kit.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.LinuxHeaders(),
				d.Libc(),
				d.Binutils(),
				d.GCC(),
				d.PkgConfig(),
				d.M4(),
				d.Automake(),
				d.Libffi(),
				d.Libtasn1(),
				Patch(d, d.P11kitSrc(), Shell(
					`cd /src/p11kit-src`,
					`sed '20,$ d' -i trust/trust-extract-compat.in`,
					`echo '/usr/libexec/make-ca/copy-trust-modifications' >> trust/trust-extract-compat.in`,
					`echo '/usr/sbin/make-ca -f -g' >> trust/trust-extract-compat.in`,
				)),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/p11kit-src/configure`,
					`--prefix=/usr`,
					`--sysconfdir=/etc`,
					`--with-trust-paths=/etc/pki/anchors`,
				}, " "),
				`make`,
				`make install`,
				`ln -sfv /usr/libexec/p11-kit/trust-extract-compat /usr/bin/update-ca-certificates`,
				`ln -sfv ./pkcs11/p11-kit-trust.so /usr/lib/libnssckbi.so`,
			),
		).With(
			Name("p11kit"),
			RuntimeDeps(
				d.Libc(),
				d.Libffi(),
				d.Libtasn1(),
			),
		).With(opts...)
	})
}
