package sysroot

import (
	"github.com/sipsma/bincastle/distro/bootstrap"
	. "github.com/sipsma/bincastle/graph"
)

type Sysroot struct{}

func (Sysroot) Spec() Spec {
	return Merge(
		Awk{},
		Bash{},
		Binutils{},
		Bison{},
		Bzip2{},
		Coreutils{},
		Diffutils{},
		File{},
		Findutils{},
		Libstdcpp{},
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
	).With(
		Unbootstrapped(bootstrap.Spec{}),
		Wrapped(AppendOutputDir("/sysroot")),
	)
}
