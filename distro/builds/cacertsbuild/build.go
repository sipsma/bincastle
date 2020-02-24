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
}, opts ...Opt) PkgBuild {
	return PkgBuildOf(d.Exec(
		p11kit.Pkg(d),
		openssl.Pkg(d),
		coreutils.Pkg(d),
		cacerts.SrcPkg(d).With(DiscardChanges()),
		Shell(
			`cd /src/cacerts-src`,
			`make -j1 install`,
			`install -vdm755 /etc/ssl/local`,
			`/usr/sbin/make-ca -g`,
		),
	).With(
		Name("cacerts"),
		VersionOf(cacerts.SrcPkg(d)),
		Deps(
			p11kit.Pkg(d),
			openssl.Pkg(d),
			coreutils.Pkg(d),
		),
	).With(opts...))
}
