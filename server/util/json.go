package util

type Json interface {
	isJson()
}

type JNull struct{}

func (JNull) isJson() {}

type JBool bool

func (JBool) isJson() {}

type JNumber float64

func (JNumber) isJson() {}

type JString string

func (JString) isJson() {}

type JArray []Json

func (JArray) isJson() {}

func JArrayFrom[T any](vals []T, conv func(T) Json) JArray {
	arr := make(JArray, len(vals))
	for i, v := range vals {
		arr[i] = conv(v)
	}
	return arr
}

type JObject map[string]Json

func (JObject) isJson() {}

func JsonEqual(a, b Json) bool {
	switch a := a.(type) {
	case JNull:
		_, ok := b.(JNull)
		return ok
	case JBool:
		v, ok := b.(JBool)
		return ok && a == v
	case JNumber:
		v, ok := b.(JNumber)
		return ok && float64(a) == float64(v)
	case JString:
		v, ok := b.(JString)
		return ok && a == v
	case JArray:
		v, ok := b.(JArray)
		if !ok || len(a) != len(v) {
			return false
		}
		for i := range a {
			if !JsonEqual(a[i], v[i]) {
				return false
			}
		}
		return true
	case JObject:
		v, ok := b.(JObject)
		if !ok || len(a) != len(v) {
			return false
		}
		for k := range a {
			if !JsonEqual(a[k], v[k]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
