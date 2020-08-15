package graph

type Metadata interface {
	Metadata(interface{}) interface{}
}

type nameKey struct{}

func NameOf(m Metadata) string {
	if m == nil {
		return ""
	}
	if v, ok := m.Metadata(nameKey{}).(string); ok {
		return v
	}
	return ""
}

// TODO maybe there should be support for multiple Names?
func Name(name string) LayerSpecOpt {
	return LayerSpecOptFunc(func(ls LayerSpecOpts) LayerSpecOpts {
		ls.SetValue(nameKey{}, name)
		return updatedOverrideEnv(ls)
	})
}

type isLocalOverrideKey struct{}

func IsLocalOverride(m Metadata) bool {
	if m == nil {
		return false
	}
	if v, ok := m.Metadata(isLocalOverrideKey{}).(bool); ok {
		return v
	}
	return false
}

type LocalOverride bool

func (o LocalOverride) ApplyToLayerSpecOpts(ls LayerSpecOpts) LayerSpecOpts {
	ls.SetValue(isLocalOverrideKey{}, bool(o))
	return updatedOverrideEnv(ls)
}

func updatedOverrideEnv(ls LayerSpecOpts) LayerSpecOpts {
	envKey := EnvOverridesPrefix+canonName(NameOf(ls))
	if !IsLocalOverride(ls) || NameOf(ls) == "" || ls.MountDir == "" {
		delete(ls.RunEnv, envKey)
		return ls
	}
	ls.RunEnv[envKey] = ls.MountDir
	return ls
}
