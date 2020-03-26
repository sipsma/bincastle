package util

type OptionKey struct{}

type OptionSet struct{
	m map[interface{}]interface{}
}

type OptionSetter interface {
	OptionSet() OptionSet
}

func (os OptionSet) GetValue(key interface{}) interface{} {
	return os.m[key]
}

func (os OptionSet) Merge(others ...OptionSetter) OptionSet {
	newM := make(map[interface{}]interface{})
	for k, v := range os.m {
		newM[k] = v
	}
	for _, other := range others {
		if other == nil {
			continue
		}
		for k, v := range other.OptionSet().m {
			newM[k] = v
		}
	}
	return OptionSet{newM}
}

func (os OptionSet) OptionSet() OptionSet {
	return os
}

func EmptyOptionSet() OptionSet {
	return OptionSet{map[interface{}]interface{}{}}
}

func Option(key interface{}, value interface{}) OptionSet {
	return OptionSet{map[interface{}]interface{}{
		key: value,
	}}
}

func GetBool(key interface{}, defaultVal bool, opts ...OptionSetter) bool {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		val, ok := opt.OptionSet().GetValue(key).(bool)
		if ok {
			return val
		}
	}
	return defaultVal
}

func GetInt(key interface{}, defaultVal int, opts ...OptionSetter) int {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		val, ok := opt.OptionSet().GetValue(key).(int)
		if ok {
			return val
		}
	}
	return defaultVal
}

func GetString(key interface{}, defaultVal string, opts ...OptionSetter) string {
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		val, ok := opt.OptionSet().GetValue(key).(string)
		if ok {
			return val
		}
	}
	return defaultVal
}
