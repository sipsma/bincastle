package src

import (
	"fmt"
	"path/filepath"

	"github.com/moby/buildkit/client/llb"
	. "github.com/sipsma/bincastle/graph"
	. "github.com/sipsma/bincastle/util"
)

type AlwaysRun bool
type alwaysRun OptionKey
func (o AlwaysRun) OptionSet() OptionSet {
	return Option(alwaysRun{}, bool(o))
}
func AlwaysRunOf(opts ...OptionSetter) bool {
	return GetBool(alwaysRun{}, false, opts...)
}

type StripComponents int
type stripComponents OptionKey
func (o StripComponents) OptionSet() OptionSet {
	return Option(stripComponents{}, int(o))
}
func StripComponentsOf(opts ...OptionSetter) int {
	return GetInt(stripComponents{}, 0, opts...)
}

type Ref string
type ref OptionKey
func (o Ref) OptionSet() OptionSet {
	return Option(ref{}, string(o))
}
func RefOf(opts ...OptionSetter) string {
	return GetString(ref{}, "master", opts...)
}

type CurlOpt struct {
	AlwaysRun
	StripComponents
}

func CurlOptionSet(opts ...CurlOpt) OptionSet {
	optionSet := EmptyOptionSet()
	for _, opt := range opts {
		optionSet = optionSet.Merge(
			opt.AlwaysRun,
			opt.StripComponents,
		)
	}
	return optionSet
}

func Curl(
	distro Executor,
	name string, // TODO make optional?
	url string,
	curlOpts ...CurlOpt,
) Pkg {
	opts := CurlOptionSet(curlOpts...)
	if name == "" {
		panic("name must be set")
	}

	if url == "" {
		panic("url must be set")
	}

	runOpts := []llb.RunOption{Shell(
		`mkdir -p /src`,
		`cd /src`,
		fmt.Sprintf("curl -L -O %s", url),
		`DLFILE=$(ls)`,
		fmt.Sprintf(
			`tar --strip-components=%d --extract --no-same-owner --file=$DLFILE`,
			StripComponentsOf(opts)),
		`rm $DLFILE`,
	)}

	if AlwaysRunOf(opts) {
		runOpts = append(runOpts, llb.IgnoreCache)
	}

	return distro.Exec(runOpts...).With(
		OutputDir("/src"),
		MountDir(filepath.Join("/src", name)),
		Name(name),
	)
}

type GitOpt struct {
	AlwaysRun
	Ref
}

func GitOptionSet(opts ...GitOpt) OptionSet {
	optionSet := EmptyOptionSet()
	for _, opt := range opts {
		optionSet = optionSet.Merge(
			opt.AlwaysRun,
			opt.Ref,
		)
	}
	return optionSet
}

func Git(
	distro Executor,
	name string,
	url string,
	gitOpts ...GitOpt,
) Pkg {
	opts := GitOptionSet(gitOpts...)
	if name == "" {
		panic("name must be set")
	}

	if url == "" {
		panic("url must be set")
	}

	runOpts := []llb.RunOption{Shell(
		`mkdir -p /src`,
		fmt.Sprintf(`git clone --recurse-submodules %s /src`, url),
		`cd /src`,
		fmt.Sprintf(`git checkout %s`, RefOf(opts)),
	)}

	if AlwaysRunOf(opts) {
		runOpts = append(runOpts, llb.IgnoreCache)
	}

	return distro.Exec(runOpts...).With(
		OutputDir("/src"),
		MountDir(filepath.Join("/src", name)),
		Name(name),
	)
}
