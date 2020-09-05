# bincastle

Bincastle is a tool for running and developing highly-portable Linux systems.

It’s a standalone binary, run as a non-root user, that will:
1. Read in a system definition from a local directory or remote git repo
1. Build the system or use cached results of previous builds (stored locally or remotely)
1. Run the system in an isolated sandbox (currently a rootless container)

Once you're in the built system you can use it however you want, but you will also still have access to bincastle. This allows you to start up new systems from within each other, including those built from locally overridden sources.

Right now the targeted use-cases are:
* **Development environments**
   * Define a highly-customized development environment, built entirely from source, and run it by just executing a standalone binary as a non-root user with [very few host requirements](#Requirements).
* **System build playgrounds**
   * Bincastle makes it trivial to modify whatever source code you want and re-build an entire system with that change, whether it’s to your code, a library, a compiler, libc, or anything else. Change and break whatever you want within the confines of an isolated sandbox and return back to a safe-harbor system once done. You can learn a lot more when the overhead of tinkering is minimized.
   
---

- [Status](#status)
  - [Current Features](#current-features)
  - [Roadmap](#roadmap)
- [Motivation](#motivation)
  - [Goals](#goals)
- [Using](#using)
  - [Requirements](#requirements)
  - [Building](#building)
  - [Demo](#demo)
    - [Starting Up](#starting-up)
    - [Adding Layers](#adding-layers)
    - [Modifying Libc](#modifying-libc)
- [Thanks](#thanks)

# Status
Very early, highly unstable alpha. Try it if you are brave but keep it away from anything you have emotional and/or financial attachment to. Many known and unknown bugs, lots of incomplete features, unnecessarily slow at times, confusing output, mostly undocumented, very few tests, etc. I use it everyday as my development environment, but it's not yet friendly enough to use as a daily driver by someone unfamiliar with the internal details.

Up to now, the work has been to explore/prove the concept and get a better handle on some of the underlying technical challenges.

Current work is focused on 
* Cleaning up and fleshing out the existing features
* Starting to follow some "more advanced" development practices such as having tests, documenting literally anything, not force-pushing to main, etc. 
* Continuing efforts to merge features upstream to Buildkit when it makes sense to do so

## Current Features
* [Very few requirements for the host system](#Requirements) in order to run and build sandboxed systems
  * Including never needing to be run as root or alongside setuid binaries
* Support for starting bincastle systems from within another in order to enable iterative development of them.
* Support for 1 system specification language and an [example using it to implement a distro](examples/distro) based on [Linux From Scratch](http://www.linuxfromscratch.org/lfs/).
  * The build specification language is a Golang library, which may seem like an odd choice of language at first. While this was initially just a practical decision due to all the other code needing to be in Go, it's actually ended up pretty low-boilerplate and simple while remaining reasonably flexible.
* Support for local and remote caching of builds (thanks to using an embedded [Buildkit](https://github.com/moby/buildkit))

[See the Demo](#Demo) to get an idea for what this all currently looks like in practice. At the moment, that Demo is the extent of the documentation 😬.

## Roadmap
"Roadmap" is a strong word right now, but in the longer-term, there's a lot more features I'd like to add to bincastle. Some of the bigger ones in no particular order:
* Integration with VMs + Kernel Development
  * Right now the sandbox used to build and run bincastle layers is a rootless container, but if support for Firecracker and/or Kata containers was added, you could use bincastle for kernel level development.
  * Running in a VM also grants true root capabilities (while still not needing root on the host, just `/dev/kvm` access), which unlocks more possibilities in terms of filesystems, networking configuration, etc.
* Support for more system specification languages. Possibilities include support for:
  * Other general-purpose languages, especially ones with more featureful type systems than Go such as Rust
  * Nix? I don't know enough about Nix internals to have a clear idea how this would work, but given the amount of effort invested in that community it would be great to find a way to make Nix an option.
  * YAML? I don't personally want this, but if there's ever interest, YAML or similar configuration languages can be supported too
* Remote migrations
  * It should be possible to migrate an instance of bincastle from one host to another without losing any state (including live running processes) via CRIU and/or VM migration depending on the backend.
* Better support for exporting build results
  * There's technically support today for exporting build results to a local dir and to container images, but they're big hacks and only used internally. These should be cleaned up and formalized.
  * Once there's support for VM-based development, it would also make sense to support exporting VM images and possibly disk images intended for bare metal targets

# Motivation
Bincastle's main goal is make it easy to develop Linux systems. By "system", I mean the rootfs of programs, libraries and other files that together create a userspace you see when you start a shell or run other executables on Linux.

To define what I mean by "developing" a system, it's worth breaking down a bit pedantically:
1. In order to run something, it needs to have been built.
2. In order to build something, you need to be able to run its pre-requisites and some build instructions.
3. In order to **develop** something, the 2 above processes need to be combined into a cohesive, iterative loop.

So, a tool for developing Linux systems should not only make (1) running and (2) building them easy, it should also (3) make the feedback between those two processes easy.

There are a lot of great existing tools like NixOS and Docker that have pushed the state of the art when it comes to running and building systems. However, I felt there was room for a new tool that filled some usability gaps, particularly when it comes to development. Bincastle is an attempt at creating a cohesive tool that keeps overhead, learning-curves and other sources of friction minimized along the whole development loop.

A cool side-effect of having a tool that makes it easy to develop systems is that you also get a tool that makes it easy to just run them if that's all you're interested in. So bincastle can be used for active system development or just as a convenient tool to run a reproducible system for other purposes.

## Goals
While enabling easy system development is the highest-level goal, it implies some important sub-goals for bincastle:
1. Bincastle should be usable from within bincastle
   * This is what makes the development process an actual loop. 
   * When you first start a system via bincastle, you bootstrap the development loop. From that initial system you can make changes to the definition of any other system (including the one you're in), rebuild and run it, which completes one iteration of that loop.
1. It should be highly-portable across Linux host systems. 
   * Running bincastle should impose as few requirements on the host system as possible. This includes dependencies on .so's and service daemons, requirements for root capabilities or setuid binaries, etc.
   * You should also be able to move your active work from one machine to another transparently.
1. It shouldn't necessarily tie you to one language for defining system builds.
   * Support for system build specification languages should be pluggable in order to cover a full spectrum of use cases, strongly-held opinions and momentary whims.
1. It should minimize the time required for each iteration of the development loop
   * It shoud thus re-use caches stored both locally and remotely whenever possible
   * In particular, if you're interested just in running a pre-built system, you should be able to just download a remote cache for it and never be forced to do any local building.
1. You should be able to use systems by just running them via bincastle, but also by exporting them to other formats
   * "Other formats" include container images, VM images, tarballs, etc.

Finally, there's a few things bincastle is (currently) **not** trying to achieve:
1. Support for using existing package managers to build systems
   * You should **always** be able to define builds of a system entirely from source. 
   * However, it's not currently expected that you can take any existing package manager like apt, yum, etc. and easily use it within bincastle. They just often operate on very different assumptions about how the system is constructed.
1. Creating *new* security boundaries
   * Sandboxing in bincastle is for the purposes of achieving portability, which requires isolation from the host system. It's not intended to create a security boundary or enhance any existing ones.
   * That being said, obviously bincastle should not enable privilege escalation or escape from any security boundaries that exist around the user at the time they run it
1. Running production software
   * Bincastle is several orders of magnitude too unstable to be considered for anything remotely near production and is currently just focused on being a development tool.

# Using
Please read the [Status](#Status) before using bincastle for anything real outside the Demo.

## Requirements
Right now, the requirements to run bincastle are:
1. x86_64 host running Linux w/ kernel v4.18 or greater
   * 4.18+ allows use of fuse from unprivileged user namespaces
   * The x86_64 requirement will go away in the future, I just haven't made builds of some bootstrap images for other architectures yet.
   * The kernel version is not likely to go down much in the future. A lot of features needed to make unprivileged user namespaces useful are only in kernels from the 4.x+ series and the pain of requiring them will subside as time goes on.
1. The `kernel.unprivileged_userns_clone` sysctl needs to be set to `1` (this is often the default setting)
1. Existence of `/dev/fuse` on the host system
   * This could be optional in the future
1. An internet connection
   * Bincastle downloads sources and/or build-cache in order to build+run systems.
   * This could be optional in the future
1. Free disk space in the filesystem your homedir is located on
   * bincastle stores all its state in `$HOME/.bincastle`, which is not configurable right now
   * `make dist-clean` will remove all local state stored by bincastle.
   * **To be safe, have at least 20 GB of space to run the full demo including rebuilding the system from scratch** (this number should be reduced in the future).

You do **not** need root to run bincastle and it's not recommended to do so (I only test it as non-root users). Don't prefix `./bincastle` with `sudo`.

## Building
Someday there'll be static builds of bincastle for you to just download and use immediately anywhere. Today is not that day, so you'll have to build it yourself in order to use it.

You'll need:
1. Golang (I've tested w/ v1.13+)
1. Standard build essentials like GCC and Make plus a static toolchain to enable linking libc statically
   * For glibc, the static library should be found in the `glibc-static` package on Fedora 32 and in the `libc6-dev` package on Ubuntu 20.04.

Then, just run `make` and a static `bincastle` binary should show up.

## Demo
This demo shows how to start up a minimal system from a definition in a git repo and then make progressively larger changes to it. 

There's some known bugs that cause ephemeral failures when downloading/building layers. You can try re-running when that happens, including with the `--verbose` flag to see the full output of each build. Feel free to open an issue on this repo for any problems you encounter either way.

### Starting Up
Once you're on a host with [the minimal pre-requisites](#Requirements) and have [built bincastle](#Building), run this command as a non-root user to start the demo system:
```
./bincastle run --import-cache eriksipsma/bincastle-demo-cache:latest https://github.com/sipsma/bincastle.git main examples/demo
```
* `--import-cache eriksipsma/bincastle-demo-cache:latest` tells bincastle where it can find cached results of previous builds so it can just download those instead of re-building everything from scratch. I've kindly uploaded a cache to a remote registry for you to use.
* `https://github.com/sipsma/bincastle.git` is the repo containing the system definition (it's the bincastle repo in this case because the demo is stored there, it could be any git repository though)
* `main` is the name of the git branch to checkout when looking for the system definition
* `examples/demo` is an optional argument that specifies the subdir containing the actual system definition. If this was left out bincastle would just use the root of the git repo.

The command may take a bit to download the remotely cached system (this should improve in the future), but once done you should see a shell prompt like:
```
bash-5.0# 
```

You're now inside the demo system running bash. There's some coreutils commands and nano available but not much else.

You may want to take a look at the definition for the system you're running. Fortunately, the source code for it is mounted right inside your homedir, so you can open it up with:
```
nano /home/user/bincastle-src/examples/demo/main.go
```

(*If you are reading these instructions without actually following them, [you can see the definition here](examples/demo/main.go)*)

### Adding Layers
Let's now try changing the system definition. There are some commented out lines that look like 
```
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
```
Try uncommenting some of them (delete the leading `//`) to add more layers you're interested in to the demo system. They are just a sample of [the layers I've currently defined for the example distro](examples/distro).

You can then exit nano, saving your changes (`ctrl-O` followed by `ctrl-X`).

The system won't immediately change, you need to rebuild it. The `bincastle` binary you used to start the system earlier on the host is mounted at `/bincastle`, so you can use it as before but with slightly different arguments this time:
```
/bincastle run --import-cache eriksipsma/bincastle-demo-cache:latest /home/user/bincastle-src examples/demo
```
* `/home/user/bincastle-src` is used because we want to use the local dir with the changes we just made instead of a remote git repo
* `examples/demo` is the same optional argument as before specifying the subdir of the repo to use.

The remote cache includes all the commented out layers, so the new system should be ready fairly quickly. Once it's done you should see a similar shell prompt as before, but now you should also have access to the programs in the layers you included.

If you want to return to the previous system, you can just exit this shell to go back (`ctrl-D` or just run the `exit` command). 

*Note: the UX around switching systems is **extremely** limited at the moment:*
* When you switch to a new system, all the changes you made to the previous one (called the "upperdir") will be overlayed onto the new system. 
   * Eventually, there will be a way to name systems+upperdirs and control which one is being used, likely alongside some sort of "branching" ability similar to but simpler than git
* When you switch to a new system, all the processes in the previous are killed, so you have to restart those
   * This will be allieviated as part of adding support for branching upperdirs (right now the problem is it's unsafe to use the same upperdir in multiple overlay mounts)
* There's not a good way of navigating the "history" of systems as you iteratively develop them. If you create a chain of them, you will need to exit each one before bincasatle will exit (basically, you'll be pressing ctrl-D a lot).
   * This will also improve as part of adding support for named systems, which should include being able to navigate+exit them more flexibly

### Modifying Libc
The last part of the demo is optional as it can take a long time to run depending on your hardware (1 hour running in a VM on a mediocre laptop, 15 minutes on a 18-core desktop). Feel free to just read along to get the gist of the idea and then play around on your own with something simpler if that's more appealing.

Open up the system definition again (using `nano` or whatever editor you added in the previous steps) and uncomment this line:
```
// Dep(Wrap(src.Libc{}, MountDir("/home/user/libc-src")))
```
This will result in the system having the source for glibc that was used to build the system mounted into it.

Rebuild the system again to get the source to actually show up:
```
/bincastle run --import-cache eriksipsma/bincastle-demo-cache:latest /home/user/bincastle-src examples/demo
```

Now that the source is mounted locally, **any future builds of any system will use this local source to build glibc**. We're going to make a change to it and rebuild the current system with it. This will take a while because libc is the "root" dependency of almost every other layer. Right now bincastle will always rebuild a layer if anything in its transitive dependencies changes, so this means almost every layer will be rebuilt (this behavior may be more configurable in the future).

To make the change, run the following command (copy the whole thing and paste it to the shell as a single command):
```
patch -u /home/user/libc-src/stdlib/exit.c <<EOF
From 58fdeb5e9104988385fa26eba6d7e7381ef1af6f Mon Sep 17 00:00:00 2001
From: Erik Sipsma <erik@sipsma.dev>
Date: Thu, 3 Sep 2020 16:24:37 -0700
Subject: [PATCH] Add process obituary.

---
 stdlib/exit.c | 4 ++++
 1 file changed, 4 insertions(+)

diff --git a/stdlib/exit.c b/stdlib/exit.c
index 49d12d9952..c4b9f4b46e 100644
--- a/stdlib/exit.c
+++ b/stdlib/exit.c
@@ -20,6 +20,7 @@
 #include <unistd.h>
 #include <sysdep.h>
 #include <libc-lock.h>
+#include <errno.h>
 #include "exit.h"

 #include "set-hooks.h"
@@ -38,6 +39,9 @@ attribute_hidden
 __run_exit_handlers (int status, struct exit_function_list **listp,
                     bool run_list_atexit, bool run_dtors)
 {
+       if (status != 0 ) {
+               fprintf(stderr, "RIP %s, who exited this world with %d.\n", program_invocation_name, status);
+       }
   /* First, call the TLS destructors.  */
 #ifndef SHARED
   if (&__call_tls_dtors != NULL)
--
2.23.0

EOF
```

The change we're making is completely useless and silly, but this is a demo after all.

Now, we're going to rebuild the system using that change. Nothing extra is necessary, but to see how it works under the hood run `env` and you'll see output like this:
```
BINCASTLE_OVERRIDE_libc_src=/home/user/libc-src
PWD=/home/user/libc-src
BINCASTLE_SOCK=/bincastle.sock
HOME=/home/user
LANG=en_US.UTF-8
TERM=xterm
SHLVL=1
PATH=/bin:/usr/bin
_=/usr/bin/env
OLDPWD=/home/user
```

The `BINCASTLE_OVERRIDE_libc_src` env var is important here, it was set by bincastle automatically when the source got mounted into the system and tells future bincastle builds to use that local source in place of the original one when doing builds.

Now, rebuild the system again with the same command as earlier:
```
/bincastle run --import-cache eriksipsma/bincastle-demo-cache:latest /home/user/bincastle-src examples/demo
```

If everything goes well, you'll see a lot of output flying by for a while as layers get built (in parallel whenever possible). However, if something goes wrong, you'll drop back to the original shell from before and can try again.

Once the build completes successfully, you'll switch to the new system. You can try out the change made to libc by running any program that exits non-zero, i.e.
```
bash-5.0# ls /foobar
ls: cannot access '/foobar': No such file or directory
RIP ls, who exited this world with 2.
bash-5.0# bash -c 'exit 123'
RIP bash, who exited this world with 123.
bash-5.0#
```

Obviously this change in particular is absurd and there's a lot of polish needed on bincastle itself, but hopefully this demo gave a sense of the long-term potential for the tool.

# Thanks
Bincastle is made possible by the shoulders of
* [Buildkit](https://github.com/moby/buildkit), which is embedded in bincastle, serves as the foundation of many of crucial features including build scheduling, logging, caching, and importing/exporting.
* [runc](https://github.com/opencontainers/runc), in particular [libcontainer](https://github.com/opencontainers/runc/tree/master/libcontainer), is what enables rootless sandboxing of systems
* [fuse-overlayfs](https://github.com/containers/fuse-overlayfs) allows overlay mounts to be created in rootless containers without any kernel patch hacks
* [Linux From Scratch](http://www.linuxfromscratch.org/lfs/)
  * Some experiments with automating Linux From Scratch using Buildkit eventually turned into very early prototypes of bincastle, and even today [the example distro](`examples/distro`) is heavily based on their instructions.
  * It's a great example of taking something that's normally inaccessible to mere mortals and bringing it within reach, which is a goal that bincastle also strives for.

Inspiration for bincastle comes from a lot of places and experiences, but a few that stick out in retrospect:
* [NixOS](https://nixos.org/)
  * NixOS really solidified in my mind the importance of defining systems using immutable definitions. Much like vertically arranged browser tabs, once you enter that mindset you question why anything else exists outside of it.
  * I hope in time there can be a way to support using Nix as one bincastle's options for system specification language
* Jess Frazelle's blog post ["Getting Towards Real Sandbox Containers"](https://blog.jessfraz.com/post/getting-towards-real-sandbox-containers/)
  * While bincastle diverges from precisely what she described in that post, it set off a lot of the ideas that led to many features in bincastle, particularly around being able to run bincastle with few host requirements and as a non-root user.
