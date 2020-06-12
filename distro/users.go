package distro

import (
	"fmt"
	"path/filepath"

	. "github.com/sipsma/bincastle/graph"
)

type User struct {
	Name    string
	Shell   string
	Homedir string
}

func (s User) Spec() Spec {
	if s.Name == "" {
		s.Name = "user"
	}
	if s.Shell == "" {
		s.Shell = "/bin/bash"
	}
	if s.Homedir == "" {
		s.Homedir = filepath.Join("/home", s.Name)
	}

	return LayerSpec(
		Dep(baseSystem{}),
		BuildDep(Coreutils{}),
		BuildOpts(),
		Shell(
			fmt.Sprintf(`echo '%s:x:0:0:%s:%s:%s' > /etc/passwd`,
				s.Name, s.Name, s.Homedir, s.Shell,
			),
			fmt.Sprintf(`echo '%s:x:0:' > /etc/group`,
				s.Name,
			),
		),
	)
}
