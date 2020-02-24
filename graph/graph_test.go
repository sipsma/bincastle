package graph

import (
	"testing"

	"github.com/moby/buildkit/client/llb"
	"github.com/stretchr/testify/require"
)

type srcAKey struct{}
func SrcA(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(srcAKey{}, p)
}

type srcBKey struct{}
func SrcB(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(srcBKey{}, p)
}

type srcCKey struct{}
func SrcC(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(srcCKey{}, p)
}

type pkg1AKey struct{}
func Pkg1A(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(pkg1AKey{}, p)
}

type pkg1BKey struct{}
func Pkg1B(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(pkg1BKey{}, p)
}

type pkg1CKey struct{}
func Pkg1C(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(pkg1CKey{}, p)
}

type pkg2BKey struct{}
func Pkg2B(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(pkg2BKey{}, p)
}

type pkg2CKey struct{}
func Pkg2C(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(pkg2CKey{}, p)
}

type pkg3BKey struct{}
func Pkg3B(c PkgCache, p func() PkgBuild) Pkg {
	return c.PkgOnce(pkg3BKey{}, p)
}

func TestTsort(t *testing.T) {
	baseSystem := DefaultPkger().Exec(
		llb.Args([]string{"mkdir", "/base"}),
	).With(Name("base-system"))

	distro := DefaultPkger(
		llb.AddEnv("PATH", "/bin"),
		baseSystem,
		AtRuntime(Deps(baseSystem)),
	)

	srcA := SrcA(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(llb.Args([]string{"get", "srcA"})).With(
			Name("srcA"),
			V(1, 2),
			AltVersions(
				distro.Exec(llb.Args([]string{"get", "srcA-v2"})).With(
					Name("srcA-v2"),
					V(2, 3),
				),
				distro.Exec(llb.Args([]string{"get", "srcA-v3"})).With(
					Name("srcA-v3"),
					V(3, 5),
				),
			),
		))
	})
	require.Equal(t, []Pkg{srcA}, srcA.Roots())
	require.Equal(t, []Pkg{baseSystem}, BuildDepsOf(srcA).Roots())
	require.Equal(t, []Pkg{baseSystem}, DepsOf(srcA).Roots())

	srcB := SrcB(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(llb.Args([]string{"get", "srcB"})).With(
			Name("srcB"),
			V(1, 4),
		))
	})
	require.Equal(t, []Pkg{srcB}, srcB.Roots())
	require.Equal(t, []Pkg{baseSystem}, BuildDepsOf(srcB).Roots())
	require.Equal(t, []Pkg{baseSystem}, DepsOf(srcB).Roots())

	srcC := SrcC(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(llb.Args([]string{"get", "srcC"})).With(
			Name("srcC"),
		))
	})
	require.Equal(t, []Pkg{srcC}, srcC.Roots())
	require.Equal(t, []Pkg{baseSystem}, BuildDepsOf(srcC).Roots())
	require.Equal(t, []Pkg{baseSystem}, DepsOf(srcC).Roots())

	pkg1A := Pkg1A(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(srcA, srcB,
			llb.Args([]string{"build", "pkg1A"}),
		).With(
			Name("pkg1A"),
			VersionOf(srcA),
		))
	})
	require.Equal(t, []Pkg{pkg1A}, pkg1A.Roots())
	require.Equal(t, []Pkg{srcB, srcA}, BuildDepsOf(pkg1A).Roots())
	require.Equal(t, []Pkg{baseSystem}, DepsOf(pkg1A).Roots())

	pkg1B := Pkg1B(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(srcB,
			llb.Args([]string{"build", "pkg1B"}),
		).With(
			Name("pkg1B"),
			VersionOf(srcB),
		))
	})
	require.Equal(t, []Pkg{pkg1B}, pkg1B.Roots())
	require.Equal(t, []Pkg{srcB}, BuildDepsOf(pkg1B).Roots())
	require.Equal(t, []Pkg{baseSystem}, DepsOf(pkg1B).Roots())

	pkg1C := Pkg1C(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(srcC,
			llb.Args([]string{"build", "pkg1C"}),
		).With(
			Name("pkg1C"),
		))
	})
	require.Equal(t, []Pkg{pkg1C}, pkg1C.Roots())
	require.Equal(t, []Pkg{srcC}, BuildDepsOf(pkg1C).Roots())
	require.Equal(t, []Pkg{baseSystem}, DepsOf(pkg1C).Roots())

	pkg2B := Pkg2B(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(
			srcC,
			pkg1B,
			llb.Args([]string{"build", "pkg2B"}),
		).With(
			Name("pkg2B"),
			Deps(pkg1A, pkg1C),
			V(2, 2),
		))
	})
	require.Equal(t, []Pkg{pkg2B}, pkg2B.Roots())
	require.Equal(t, []Pkg{
		pkg1B,
		srcC,
	}, BuildDepsOf(pkg2B).Roots())
	require.Equal(t, []Pkg{pkg1C, pkg1A}, DepsOf(pkg2B).Roots())

	srcAVersion := srcA.With(SwapToVersion(V(2, 3)))
	pkg2C := Pkg2C(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(
			srcAVersion,
			pkg1A,
			pkg1B,
			llb.Args([]string{"build", "pkg2C"}),
		).With(
			Name("pkg2C"),
			Deps(pkg1B, pkg1C),
			V(2, 3),
		))
	})
	require.Equal(t, []Pkg{pkg2C}, pkg2C.Roots())
	require.Equal(t, []Pkg{
		srcAVersion,
		pkg1B,
		pkg1A,
	}, BuildDepsOf(pkg2C).Roots())
	require.Equal(t, []Pkg{pkg1C, pkg1B}, DepsOf(pkg2C).Roots())

	srcAVersion = srcA.With(SwapToVersion(V(3, 5)))
	pkg3B := Pkg3B(distro, func() PkgBuild {
		return PkgBuildOf(distro.Exec(
			srcAVersion,
			llb.Args([]string{"build", "pkg3B"}),
		).With(
			Name("pkg3B"),
			Deps(
				pkg1A,
				pkg2C,
				pkg2B,
			),
			VersionOf(srcAVersion),
		))
	})
	require.Equal(t, []Pkg{pkg3B}, pkg3B.Roots())
	require.Equal(t, []Pkg{
		srcAVersion,
	}, BuildDepsOf(pkg3B).Roots())
	require.Equal(t, []Pkg{pkg2B, pkg2C}, DepsOf(pkg3B).Roots())

	// verify trying to repackage doesn't change anything
	pkg3B = Pkg3B(distro, func() PkgBuild {
		return PkgBuildOf(EmptyPkg())
	})

	tsorted := Tsort(pkg3B)
	var names []string
	var versions []Version
	for _, pkg := range tsorted {
		names = append(names, NameOf(pkg))
		versions = append(versions, VersionOf(pkg))
	}

	require.Equal(t, []string{
		"pkg3B",
		"pkg2B",
		"pkg2C",
		"pkg1C",
		"pkg1B",
		"pkg1A",
		"base-system",
	}, names)

	require.Equal(t, []Version{
		V(3, 5),
		V(2, 2),
		V(2, 3),
		NoVersion,
		V(1, 4),
		V(1, 2),
		NoVersion,
	}, versions)
}
