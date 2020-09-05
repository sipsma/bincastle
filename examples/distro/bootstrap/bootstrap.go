package bootstrap

import (
	"github.com/sipsma/bincastle/graph"
)

func BuildOpts() graph.LayerSpecOpt {
	return graph.MergeLayerSpecOpts(
		graph.BuildEnv("LC_ALL", "POSIX"),
		graph.BuildEnv("FORCE_UNSAFE_CONFIGURE", "1"),
		// TODO support putting env in run opts and set default $PATH via that
		graph.BuildEnv("PATH", "/tools/bin:/bin:/usr/bin"),
	)
}

type Spec struct {}

func (Spec) Spec() graph.Spec {
	return graph.Wrap(
		graph.Image{Ref: "docker.io/eriksipsma/bincastle-sysroot:latest"},
		graph.AppendOutputDir("/sysroot")).Spec()
}

// TODO figure out how to unify this with the above bootstrap (so sysroot is
// its own bootstrap).
type SysrootBootstrap struct{}

func (SysrootBootstrap) Spec() graph.Spec {
	return graph.LayerSpec(
		graph.Dep(graph.Image{Ref: "docker.io/eriksipsma/bincastle-bootstrap:latest"}),
		BuildOpts(),
		graph.BuildScript(
			`ln -sv /sysroot/tools /`,
			`mkdir -pv /sysroot/tools`,
			`mkdir -v /tools/lib`,
			`ln -sv lib /tools/lib64`,
		),
	)
}
