package distro

import (
	"strings"

	"github.com/sipsma/bincastle/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Emacs struct{}

func (Emacs) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Ncurses{}),
		Dep(Zlib{}),
		Dep(Acl{}),
		Dep(Attr{}),
		Dep(GMP{}),
		Dep(Libffi{}),
		Dep(GNUTLS{}),
		Dep(CACerts{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Emacs{}),
		BuildOpts(),
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
	)
}
