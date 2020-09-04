package sysroot

import (
	"github.com/sipsma/bincastle/examples/distro/bootstrap"
	. "github.com/sipsma/bincastle/graph"
)

type Sysroot struct{}

func (Sysroot) Spec() Spec {
	// TODO the stripping here only saves space due to
	// the fact that bincastle's image export currently
	// compresses all layers into a single layer in the final
	// container image. Once that hack is cleaned up, there
	// will need to be a way of doing this disk saving within
	// each layer build.
	return LayerSpec(
		BuildDep(bootstrap.SysrootBootstrap{}),
		Dep(Merge(
			Awk{},
			Bash{},
			Binutils{},
			Bison{},
			Bzip2{},
			Coreutils{},
			Diffutils{},
			File{},
			Findutils{},
			GCC{},
			Gettext{},
			Grep{},
			Gzip{},
			Libc{},
			LinuxHeaders{},
			M4{},
			Make{},
			Ncurses{},
			Patch{},
			Perl5{},
			Python3{},
			Sed{},
			Tar{},
			Texinfo{},
			Xz{},
			Golang{},
			Git{},
			Curl{},
			OpenSSH{},
		).With(
			Replaced(bootstrap.Spec{}, nil),
			Replaced(tmpBinutils{}, nil),
			Replaced(tmpGCC{}, nil),
			Replaced(libstdcpp{}, nil),
		)),
		bootstrap.BuildOpts(),
		BuildScript(
			`/usr/bin/strip --strip-unneeded $(find /sysroot/tools -type f -executable -exec sh -c "file -i '{}' | grep -q 'x-executable; charset=binary'" \; -print)`,
			`/usr/bin/strip --strip-debug $(find /sysroot/tools -type f -executable -exec sh -c "file -i '{}' | grep -q 'x-executable; charset=binary'" \; -print)`,
			`/usr/bin/strip --strip-debug $(find /sysroot/tools -type f -name \*.so\* -exec sh -c "file -i '{}' | grep -q 'x-sharedlib; charset=binary'" \; -print)`,
			`/usr/bin/strip --strip-debug $(find /sysroot/tools -type f -name \*.a -exec sh -c "file -i '{}' | grep -q 'x-archive; charset=binary'" \; -print)`,
			`rm -rf /sysroot/tools/{,share}/{info,man,doc}`,
			`find /sysroot/tools/{lib,libexec} -name \*.la -delete`,
			`rm -rf /src`,
			`rm /sysroot/tools/lib64`,
			`/bin/ln -sv lib /sysroot/tools/lib64`,

			`mkdir -p /sysroot/{bin,usr/{bin,lib}}`,
			`ln -sv /tools/bin/{bash,cat,chmod,dd,echo,ln,mkdir,pwd,rm,stty,touch} /sysroot/bin`,
			`ln -sv bash /sysroot/bin/sh`,
			`ln -sv /tools/bin/{env,install,perl,printf} /sysroot/usr/bin`,
			`ln -sv /tools/lib/libgcc_s.so{,.1} /sysroot/usr/lib`,
			`ln -sv /tools/lib/libstdc++.{a,so{,.6}} /sysroot/usr/lib`,
		),
	).With(
		Wrapped(AppendOutputDir("/sysroot")),
	)
}
