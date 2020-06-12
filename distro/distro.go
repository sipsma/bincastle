package distro

import (
	"runtime"
	"strconv"

	"github.com/sipsma/bincastle/distro/bootstrap"
	"github.com/sipsma/bincastle/distro/sysroot"
	. "github.com/sipsma/bincastle/graph"
)

func BuildOpts() LayerSpecOpt {
	return MergeLayerSpecOpts(
		// TODO should this really be hardcoded?
		// maybe something smaller than "all cpus" would be a better default
		Env("MAKEFLAGS", "-j"+strconv.Itoa(runtime.NumCPU())),
		Env("LC_ALL", "POSIX"),
		Env("FORCE_UNSAFE_CONFIGURE", "1"), // builds think we are "root" (really just in unpriv userns)
		// TODO support putting env in run opts and set default $PATH via that
		Env("PATH", "/bin:/usr/bin:/sbin:/usr/sbin:/tools/bin"),
	)
}

type baseSystem struct{}

func (baseSystem) Spec() Spec {
	return LayerSpec(
		BuildDep(sysroot.Sysroot{}),
		bootstrap.BuildOpts(),
		Shell(
			`mkdir -pv /{bin,boot,etc/{opt,sysconfig},home,lib/firmware,mnt,opt}`,
			`mkdir -pv /{media/{floppy,cdrom},sbin,srv,var}`,
			`install -dv -m 0750 /root`,
			`install -dv -m 1777 /tmp /var/tmp`,
			`mkdir -pv /usr/{,local/}{bin,include,lib,sbin,src}`,
			`mkdir -pv /usr/{,local/}share/{color,dict,doc,info,locale,man}`,
			`mkdir -v  /usr/{,local/}share/{misc,terminfo,zoneinfo}`,
			`mkdir -v  /usr/libexec`,
			`mkdir -pv /usr/{,local/}share/man/man{1..8}`,
			`mkdir -v  /usr/lib/pkgconfig`,
			`mkdir -v /lib64`,
			`mkdir -v /var/{log,mail,spool}`,
			`ln -sv /run /var/run`,
			`ln -sv /run/lock /var/lock`,
			`mkdir -pv /var/{opt,cache,lib/{color,misc,locate},local}`,
			`ln -sv bash /bin/sh`,
			`ln -sv /proc/self/mounts /etc/mtab`,
		),
	)
}

type unpatchedBaseSystem struct{}

func (unpatchedBaseSystem) Spec() Spec {
	return LayerSpec(
		Dep(sysroot.Sysroot{}),
		Dep(baseSystem{}),
		bootstrap.BuildOpts(),
		Shell(
			`ln -sv /tools/bin/{bash,cat,chmod,dd,echo,ln,mkdir,pwd,rm,stty,touch} /bin`,
			`ln -sv /tools/bin/{env,install,perl,printf} /usr/bin`,
			`ln -sv /tools/lib/libgcc_s.so{,.1} /usr/lib`,
			`ln -sv /tools/lib/libstdc++.{a,so{,.6}} /usr/lib`,
		),
	)
}

type patchedBaseSystem struct{}

func (patchedBaseSystem) Spec() Spec {
	return LayerSpec(
		Dep(unpatchedBaseSystem{}),
		bootstrap.BuildOpts(),
		Shell(
			`mv -v /tools/bin/{ld,ld-old}`,
			`mv -v /tools/$(uname -m)-pc-linux-gnu/bin/{ld,ld-old}`,
			`mv -v /tools/bin/{ld-new,ld}`,
			`ln -sv /tools/bin/ld /tools/$(uname -m)-pc-linux-gnu/bin/ld`,
			`gcc -dumpspecs | sed -e 's@/tools@@g' -e '/\*startfile_prefix_spec:/{n;s@.*@/usr/lib/ @}'  -e '/\*cpp:/{n;s@$@ -isystem /usr/include@}' > $(dirname $(gcc --print-libgcc-file-name))/specs`,
		),
	)
}

func BuildDistro(specs ...AsSpec) *Graph {
	return Build(Merge(specs...).With(Unbootstrapped(
		unpatchedBaseSystem{},
		patchedBaseSystem{},
	)))
}
