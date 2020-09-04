package sysroot

import (
	"strings"

	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

type Golang struct{}

func (Golang) Spec() Spec {
	return LayerSpec(
		Dep(Libc{}),
		BuildDep(bootstrap.Spec{}),
		BuildDep(LinuxHeaders{}),
		BuildDep(GCC{}),
		BuildDep(Binutils{}),
		BuildDep(src.Golang{}),
		BuildScratch(`/build`),
		bootstrap.BuildOpts(),
		BuildScript(
			`cd /src/golang-bootstrap-src/src`,
			`./make.bash`,
			`cd /src/golang-src/src`,
			strings.Join([]string{
				`GOROOT_BOOTSTRAP=/src/golang-bootstrap-src`,
				`GOROOT_FINAL=/tools/lib/go`,
				`GOBIN=/tools/bin`,
				`./make.bash`,
			}, " "),
			`mkdir -p /tools/lib/go`,
			`mv /src/golang-src/* /tools/lib/go`,
			`ln -s /tools/lib/go/bin/go /tools/bin/go`,
			`find /tools/lib/go -type d -name testdata -exec rm -rf "{}" \; || true`,
		),
	)
}
