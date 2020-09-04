package distro

import (
	"github.com/sipsma/bincastle/cmd"
	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	. "github.com/sipsma/bincastle/graph"
)

func BuildOpts() LayerSpecOpt {
	return MergeLayerSpecOpts(
		BuildEnv("LC_ALL", "POSIX"),
		BuildEnv("FORCE_UNSAFE_CONFIGURE", "1"), // builds think we are "root" (really just in unpriv userns)
		// TODO support putting env in run opts and set default $PATH via that
		BuildEnv("PATH", "/bin:/usr/bin:/sbin:/usr/sbin:/tools/bin"),
	)
}

type baseSystem struct{}

func (baseSystem) Spec() Spec {
	return LayerSpec(
		Dep(bootstrap.Spec{}),
		bootstrap.BuildOpts(),
		BuildScript(
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
			`ln -sv /proc/self/mounts /etc/mtab`,
		),
	)
}

type patchedBaseSystem struct{}

func (patchedBaseSystem) Spec() Spec {
	return LayerSpec(
		Dep(baseSystem{}),
		bootstrap.BuildOpts(),
		BuildScript(
			`mv -v /tools/bin/{ld,ld-old}`,
			`mv -v /tools/$(uname -m)-pc-linux-gnu/bin/{ld,ld-old}`,
			`mv -v /tools/bin/{ld-new,ld}`,
			`ln -sv /tools/bin/ld /tools/$(uname -m)-pc-linux-gnu/bin/ld`,
			`gcc -dumpspecs | sed -e 's@/tools@@g' -e '/\*startfile_prefix_spec:/{n;s@.*@/usr/lib/ @}'  -e '/\*cpp:/{n;s@$@ -isystem /usr/include@}' > $(dirname $(gcc --print-libgcc-file-name))/specs`,
		),
	)
}

func Distro(opts ...LayerSpecOpt) AsSpec {
	return LayerSpec(opts...).With(
		Replaced(patchedBaseSystem{}, baseSystem{}),
		Replaced(bootstrap.Spec{}, nil),
		EnvOverrides{},
	)
}

func WriteSystemDef(opts ...LayerSpecOpt) {
	cmd.WriteSystemDef(Distro(opts...))
}
