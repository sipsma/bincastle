package usersbuild

import (
	"fmt"

	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"

	"github.com/sipsma/bincastle/distro/pkgs/coreutils"
	"github.com/sipsma/bincastle/distro/pkgs/users"
)

func SingleUser(d interface {
		PkgCache
		Executor
		coreutils.Pkger
	},
	rootUsername string,
	homeDir string,
	shell string,
	opts ...Opt,
) users.Pkg {
	return users.BuildPkg(d, func() Pkg {
		return d.Exec(
			BuildDeps(
				d.Coreutils(),
			),
			Shell(
				fmt.Sprintf(`echo '%s:x:0:0:%s:%s:%s' > /etc/passwd`,
					rootUsername, rootUsername, homeDir, shell,
				),
				fmt.Sprintf(`echo '%s:x:0:' > /etc/group`,
					rootUsername,
				),
				fmt.Sprintf(`mkdir -p %s`, homeDir),
			),
		).With(
			Name("users"),
		).With(opts...)
	})
}
