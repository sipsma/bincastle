package cacertsbuild

import (
	"github.com/sipsma/bincastle/distro/pkgs/coreutils"
	"github.com/sipsma/bincastle/distro/pkgs/cacerts"
	"github.com/sipsma/bincastle/distro/pkgs/openssl"
	"github.com/sipsma/bincastle/distro/pkgs/p11kit"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	cacerts.Srcer
	p11kit.Pkger
	openssl.Pkger
	coreutils.Pkger
}, opts ...Opt) cacerts.Pkg {
	return cacerts.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.P11kit(),
				d.OpenSSL(),
				d.Coreutils(),
				d.CACertsSrc().With(DiscardChanges()),
			),
			Shell(
				`cd /src/cacerts-src`,
				`make -j1 install`,
				`install -vdm755 /etc/ssl/local`,
				`/usr/sbin/make-ca -g`,
			),
		).With(
			Name("cacerts"),
			RuntimeDeps(
				d.P11kit(),
				d.OpenSSL(),
				d.Coreutils(),
			),
		).With(opts...)
	})
}
