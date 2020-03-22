package emacsbuild

import (
	"strings"

	"github.com/sipsma/bincastle/distro/pkgs/acl"
	"github.com/sipsma/bincastle/distro/pkgs/attr"
	"github.com/sipsma/bincastle/distro/pkgs/cacerts"
	"github.com/sipsma/bincastle/distro/pkgs/emacs"
	"github.com/sipsma/bincastle/distro/pkgs/gnutls"
	"github.com/sipsma/bincastle/distro/pkgs/gcc"
	"github.com/sipsma/bincastle/distro/pkgs/gmp"
	"github.com/sipsma/bincastle/distro/pkgs/libc"
	"github.com/sipsma/bincastle/distro/pkgs/libffi"
	"github.com/sipsma/bincastle/distro/pkgs/linux"
	"github.com/sipsma/bincastle/distro/pkgs/ncurses"
	"github.com/sipsma/bincastle/distro/pkgs/pkgconfig"
	"github.com/sipsma/bincastle/distro/pkgs/zlib"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

func Default(d interface {
	PkgCache
	Executor
	emacs.Srcer
	libc.Pkger
	linux.HeadersPkger
	gcc.Pkger
	pkgconfig.Pkger
	ncurses.Pkger
	zlib.Pkger
	acl.Pkger
	attr.Pkger
	gmp.Pkger
	libffi.Pkger
	gnutls.Pkger
	cacerts.Pkger
}, opts ...Opt) emacs.Pkg {
	return emacs.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.Libc(),
				d.LinuxHeaders(),
				d.GCC(),
				d.PkgConfig(),
				d.Ncurses(),
				d.Zlib(),
				d.Acl(),
				d.Attr(),
				d.GMP(),
				d.Libffi(),
				d.GNUTLS(),
				d.CACerts(),
				d.EmacsSrc(),
			),
			ScratchMount(`/build`),
			Shell(
				`cd /build`,
				strings.Join([]string{
					`/src/emacs-src/configure`,
					`--prefix=/usr`,
					`--localstatedir=/var`,
					`--with-gif=no`,
					`--with-tiff=no`,
				}, " "),
				`make`,
				`make install`,
				`chown -v -R 0:0 /usr/share/emacs/26.3`,
				`rm -vf /usr/lib/systemd/user/emacs.service`,
			),
		).With(
			Name("emacs"),
			RuntimeDeps(
				d.Libc(),
				d.Ncurses(),
				d.Zlib(),
				d.Acl(),
				d.Attr(),
				d.GMP(),
				d.Libffi(),
				d.GNUTLS(),
				d.CACerts(),
			),
		).With(opts...)
	})
}
