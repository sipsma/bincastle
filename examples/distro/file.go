package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type File struct{}

func (File) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(Zlib{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(src.File{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			`cd /build`,
			strings.Join([]string{`/src/file-src/configure`,
				`--prefix=/usr`,
			}, " "),
			`make`,
			`make install`,
		),
	)
}
