package bootstrap

import (
	"runtime"
	"strconv"

	"github.com/sipsma/bincastle/graph"
)

func BuildOpts() graph.LayerSpecOpt {
	return graph.MergeLayerSpecOpts(
		// TODO should this really be hardcoded?
		// maybe something smaller than "all cpus" would be a better default
		graph.Env("MAKEFLAGS", "-j"+strconv.Itoa(runtime.NumCPU())),
		graph.Env("LC_ALL", "POSIX"),
		graph.Env("FORCE_UNSAFE_CONFIGURE", "1"),
		// TODO support putting env in run opts and set default $PATH via that
		graph.Env("PATH", "/tools/bin:/bin:/usr/bin"),
	)
}

type Spec struct{}

func (Spec) Spec() graph.Spec {
	return graph.LayerSpec(
		graph.Dep(graph.Image{Ref: "docker.io/eriksipsma/bincastle-bootstrap:latest"}),
		BuildOpts(),
		graph.Shell(
			`ln -sv /sysroot/tools /`,
			`mkdir -pv /sysroot/tools`,
		),
	)
}
