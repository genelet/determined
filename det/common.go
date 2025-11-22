package det

import (
	"fmt"
	utf8 "unicode/utf8"
)

// NewValue constructs a Value from generic Go interface v.
//
//	╔═══════════════════════════╤══════════════════════════════╗
//	║ Go type                   │ Conversion                   ║
//	╠═══════════════════════════╪══════════════════════════════╣
//	║ string                    │ ending SingleStruct value    ║
//	║ []string                  │ ending ListStruct value      ║
//	║ map[string]string         │ ending MapStruct value       ║
//	║                           │                              ║
//	║ [2]any            │ SingleStruct value           ║
//	║ [][2]any          │ ListStruct value             ║
//	║ map[string][2]any │ MapStruct value              ║
//	║                           │                              ║
//	║ *Struct                   │ SingleStruct                 ║
//	║ []*Struct                 │ ListStruct                   ║
//	║ map[string]*Struct        │ MapStruct                    ║
//	╚═══════════════════════════╧══════════════════════════════╝
func NewValue(v any) (*Value, error) {

	singleValue := func(x [2]any) (*Value, error) {
		y, err := newSingleStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: y}}, nil
	}
	listValue := func(x [][2]any) (*Value, error) {
		y, err := newListStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_ListStruct{ListStruct: y}}, nil
	}
	mapValue := func(x map[string][2]any) (*Value, error) {
		y, err := newMapStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_MapStruct{MapStruct: y}}, nil
	}

	switch t := v.(type) {
	case string:
		x := [2]any{t}
		return singleValue(x)
	case [2]any:
		return singleValue(t)
	case []string:
		var x [][2]any
		for _, u := range t {
			x = append(x, [2]any{u})
		}
		return listValue(x)
	case [][2]any:
		return listValue(t)
	case map[string]string:
		x := make(map[string][2]any)
		for k, s := range t {
			x[k] = [2]any{s}
		}
		return mapValue(x)
	case map[string][2]any:
		return mapValue(t)
	case *Struct:
		return &Value{Kind: &Value_SingleStruct{SingleStruct: t}}, nil
	case []*Struct:
		v := &ListStruct{ListFields: t}
		return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case map[string]*Struct:
		v := &MapStruct{MapFields: t}
		return &Value{Kind: &Value_MapStruct{MapStruct: v}}, nil
	default:
	}
	return nil, fmt.Errorf("invalid type: %T", v)
}

// NewStruct constructs a Struct from a generic Go map.
// v is optionally map[string]any.
// The map values are converted using NewValue.
func NewStruct(name string, v ...map[string]any) (*Struct, error) {
	x := &Struct{ClassName: name}
	if len(v) == 0 {
		return x, nil
	}

	x.Fields = make(map[string]*Value)
	for _, m := range v {
		for key, val := range m {
			if !utf8.ValidString(key) {
				return nil, fmt.Errorf("invalid UTF-8 in key: %q", key)
			}
			var err error
			x.Fields[key], err = NewValue(val)
			if err != nil {
				return nil, err
			}
		}
	}
	return x, nil
}

func newSingleStruct(v [2]any) (*Struct, error) {
	name, ok := v[0].(string)
	if !ok {
		return nil, fmt.Errorf("class name has to be string %T", v[0])
	}

	if v[1] == nil {
		return NewStruct(name)
	}
	hash, ok := v[1].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("the second has to be map[string]any %T", v[1])
	}

	return NewStruct(name, hash)
}

func newListStruct(v [][2]any) (*ListStruct, error) {
	var err error
	x := make([]*Struct, len(v))
	for i, u := range v {
		x[i], err = newSingleStruct(u)
		if err != nil {
			return nil, err
		}
	}
	return &ListStruct{ListFields: x}, nil
}

func newMapStruct(v map[string][2]any) (*MapStruct, error) {
	var err error
	x := make(map[string]*Struct)
	for i, u := range v {
		x[i], err = newSingleStruct(u)
		if err != nil {
			return nil, err
		}
	}
	return &MapStruct{MapFields: x}, nil
}
