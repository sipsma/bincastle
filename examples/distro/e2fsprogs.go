package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type E2fsprogs struct{}

func (E2fsprogs) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(UtilLinux{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(PkgConfig{}),
		BuildDep(Gzip{}),
		BuildDep(Texinfo{}),
		BuildDep(src.E2fsprogs{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{
				`/src/e2fsprogs-src/configure`,
				`--prefix=/usr`,
				`--bindir=/bin`,
				`--with-root-prefix=""`,
				`--enable-elf-shlibs`,
				`--disable-libblkid`,
				`--disable-libuuid`,
				`--disable-uuidd`,
				`--disable-fsck`,
			}, " "),
			`make`,
			`make install`,
			`make install-libs`,
			`chmod -v u+w /usr/lib/{libcom_err,libe2p,libext2fs,libss}.a`,
			`gunzip -v /usr/share/info/libext2fs.info.gz`,
			`install-info --dir-file=/usr/share/info/dir /usr/share/info/libext2fs.info`,
		),
	)
}
