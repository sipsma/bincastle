package distro

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Golang struct{}

func (Golang) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		Dep(CACerts{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(src.Golang{}),
		BuildOpts(),
		BuildScratch(`/build`),
		BuildScript(
			// TODO does this leave anything under /src ?
			`cd /src/golang-bootstrap-src/src`,
			`./make.bash`,
			`cd /src/golang-src/src`,
			strings.Join([]string{
				`GOROOT_BOOTSTRAP=/src/golang-bootstrap-src`,
				`GOROOT_FINAL=/usr/lib/go`,
				`GOBIN=/usr/bin`,
				`./make.bash`,
			}, " "),
			`mkdir -p /usr/lib/go`,
			`mv /src/golang-src/* /usr/lib/go`,
			`ln -s /usr/lib/go/bin/go /usr/bin/go`,
		),
	)
}
