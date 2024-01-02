package dethcl

import (
	utf8 "unicode/utf8"

	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type Labels []string

// NewValue constructs a Value from generic Go interface v.
//
//	╔═════════════════════════════╤══════════════════════════════╗
//	║ Go type                     │ Conversion                   ║
//	╠═════════════════════════════╪══════════════════════════════╣
//	║ string                      │ ending SingleStruct value    ║
//	║ []string                    │ ending ListStruct value      ║
//	║ map[string]string           │ ending MapStruct value       ║
//	║ map[[]string]string         │ ending HclStruct value       ║
//	║                             │                              ║
//	║ [2]interface{}              │ SingleStruct value           ║
//	║ [][2]interface{}            │ ListStruct value             ║
//	║ map[string][2]interface{}   │ MapStruct value              ║
//	║ map[[]string][2]interface{} │ HclStruct value              ║
//	║                             │                              ║
//	║ *Struct                     │ SingleStruct                 ║
//	║ []*Struct                   │ ListStruct                   ║
//	║ map[string]*Struct          │ MapStruct                    ║
//	║ map[[]string]*Struct        │ HclStruct                    ║
//	╚═════════════════════════════╧══════════════════════════════╝
func NewValue(v interface{}) (*Value, error) {

	singleValue := func(x [2]interface{}) (*Value, error) {
		y, err := newSingleStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: y}}, nil
	}
	listValue := func(x [][2]interface{}) (*Value, error) {
		y, err := newListStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_ListStruct{ListStruct: y}}, nil
	}
	mapValue := func(x map[string][2]interface{}) (*Value, error) {
		y, err := newMapStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_MapStruct{MapStruct: y}}, nil
	}
	hclValue := func(x map[Labels][2]interface{}) (*Value, error) {
		y, err := newHclStruct(x)
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_HclStruct{HclStruct: y}}, nil
	}

	switch t := v.(type) {
	case string:
		x := [2]interface{}{t}
		return singleValue(x)
	case [2]interface{}:
		return singleValue(t)
	case []string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u})
		}
		return listValue(x)
	case [][2]interface{}:
		return listValue(t)
	case map[string]string:
		x := make(map[string][2]interface{})
		for k, s := range t {
			x[k] = [2]interface{}{s}
		}
		return mapValue(x)
	case map[string][2]interface{}:
		return mapValue(t)
	case map[[]string]string:
		x := make(map[[]string][2]interface{})
		for k, s := range t {
			x[k] = [2]interface{}{s}
		}
		return mapValue(x)
	case map[[]string][2]interface{}:
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
	return nil, protoimpl.X.NewError("invalid type: %T", v)
}

// NewStruct constructs a Struct from a generic Go map.
// v is optionally map[string]interface{}.
// The map values are converted using NewValue.
func NewStruct(name string, v ...map[string]interface{}) (*Struct, error) {
	x := &Struct{ClassName: name}
	if v == nil {
		return x, nil
	}

	x.Fields = make(map[string]*Value)
	for key, val := range v[0] {
		if !utf8.ValidString(key) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in key: %q", key)
		}
		var err error
		x.Fields[key], err = NewValue(val)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

func newSingleStruct(v [2]interface{}) (*Struct, error) {
	name, ok := v[0].(string)
	if !ok {
		return nil, protoimpl.X.NewError("class name has to be string %T", v[0])
	}

	if v[1] == nil {
		return NewStruct(name)
	}
	hash, ok := v[1].(map[string]interface{})
	if !ok {
		return nil, protoimpl.X.NewError("the second has to be map[string]interface{} %T", v[1])
	}

	return NewStruct(name, hash)
}

func newListStruct(v [][2]interface{}) (*ListStruct, error) {
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

func newMapStruct(v map[string][2]interface{}) (*MapStruct, error) {
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
