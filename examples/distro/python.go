package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Python3 struct{}

func (Python3) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		Dep(Bzip2{}),
		Dep(Libffi{}),
		Dep(Ncurses{}),
		Dep(GDBM{}),
		Dep(Expat{}),
		Dep(OpenSSL{}),
		Dep(Xz{}),
		Dep(Readline{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(Binutils{}),
		BuildDep(GCC{}),
		BuildDep(PkgConfig{}),
		BuildDep(src.Python3{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/python3-src/configure`,
				`--prefix=/usr`,
				`--enable-shared`,
				`--with-system-expat`,
				`--with-system-ffi`,
				`--with-ensurepip=yes`,
			}, " "),
			`make`,
			`make install`,
			`chmod -v 755 /usr/lib/libpython3.7m.so`,
			`chmod -v 755 /usr/lib/libpython3.so`,
			`ln -sfv pip3.7 /usr/bin/pip3`,
			`ln -sv /usr/bin/python3 /usr/bin/python`,
		),
	)
}
