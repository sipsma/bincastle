package main

import (
	"github.com/sipsma/bincastle/examples/distro"
	"github.com/sipsma/bincastle/examples/distro/src"
	. "github.com/sipsma/bincastle/graph"
)

func main() {
	// distro.WriteSystemDef is a util function that takes a list
	// of layer options, turns it into a spec for a full distro and
	// writes that definition to stdout. This allows bincastle to
	// read that definition by executing this main func as a child
	// process.
	distro.WriteSystemDef(
		// All of the args to WriteSystemDef are of type
		// "LayerSpecOpt". They are combined together to create a
		// layer, which is basically just a filesystem diff.

		// "Dep" says that when this layer is built and run, the
		// arg to Dep should be present as a layer underneath it.
		// Defining Deps between layers allows you to create a DAG
		// of dependencies. When you run a layer, the DAG is sorted
		// topologically and turned into an overlay mount, with all
		// the deps as lowerdirs and the new layer being built as
		// the upperdir.
		//
		// Here, "Dep" is used, but you can also specify "BuildDep",
		// which means the dep layer will only be present when the
		// this layer is built, or "RunDep", which means the dep layer
		// will instead only be present when this layer is run after
		// build. "Dep" is just a shorthand for adding something as
		// both a BuildDep and a RunDep; it's a good default unless
		// you are sure something only needs to be present at one or
		// the other.
		Dep(distro.Coreutils{}),
		Dep(distro.Bash{}),
		Dep(distro.Nano{}),
		Dep(distro.Patch{}),
		Dep(distro.User{
			Name:    "user",
			Shell:   "/bin/bash",
			Homedir: "/home/user",
		}),

		// Coreutils, Bash and Nano give you a very basic usable shell
		// and editor, but you can add more layers whenever you want.
		// Bincastle will only build the new parts of the DAG and re-use
		// as much as possible from previous builds.
		//
		// Here's some examples of programs you can try adding by
		// uncommenting them below:
		// Dep(distro.Vim{}),
		// Dep(distro.Emacs{}),
		// Dep(distro.Tmux{}),
		// Dep(distro.Git{}),
		// Dep(distro.Golang{}),
		// Dep(distro.Python3{}),
		// Dep(distro.Procps{}),
		// Dep(distro.Which{}),
		// Dep(distro.Curl{}),
		// Dep(distro.GCC{}),

		// More layers are available under examples/distro. You can
		// also define your own wherever you want and include them here.

		// This Dep mounts this bincastle repo within the user's homedir,
		// allowing you to edit the system definition from within bincastle
		// and try out a new system.
		Dep(Wrap(src.ViaGit{
			URL: "https://github.com/sipsma/bincastle.git",
			Ref:  "main",
			Name: "bincastle-src",
		}, MountDir("/home/user/bincastle-src"))),

		// Uncomment the line below to include the libc source code used
		// to build the system under the user's home dir. You can try
		// editing the source and rebuild the system (using the binary
		// mounted at /bincastle) to try out a new system rebuilt using
		// your libc change. Note: this will take a long time as libc is
		// the "root" of the system; everything will need to be. rebuilt.

		// Dep(Wrap(src.Libc{}, MountDir("/home/user/libc-src"))),

		// These options tell bincastle what to do when this layer is
		// run as an exec via "bincastle run ...". This layer is unusual
		// in that there are no BuildArgs (or BuildScript), which you will
		// see in other layer definitions. This is because there's actually
		// nothing to really build here; this layer just needs to specify
		// deps and runtime args rather than create any new files of its own.
		Env("PATH", "/bin:/usr/bin"),
		Env("TERM", "xterm"),
		Env("LANG", "en_US.UTF-8"),
		RunWorkingDir("/home/user"),
		RunArgs("/bin/bash", "-l"),
	)
}
