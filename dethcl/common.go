package dethcl

import (
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	utf8 "unicode/utf8"
)

// NewValue constructs a Value from generic Go interface v.
//
//	╔══════════════════╤═══════════════════╗
//	║ Go type          │ Conversion        ║
//	╠══════════════════╪═══════════════════╣
//	║ string           │ ending Struct     ║
//	║ [2]interface{}   │ SingleStruct      ║
//  ║                  │                   ║
//	║ []string         │ ending ListStruct ║
//	║ [][2]interface{} │ ListStruct        ║
//  ║                  │                   ║
//  ║ *Struct          │ SingleStruct      ║
//  ║ []*Struct        │ ListStruct        ║
//	╚══════════════════╧═══════════════════╝
func NewValue(v interface{}) (*Value, error) {
	switch t := v.(type) {
	// string is treated as ending Struct without fields
	case string:
		v, err := newSingleStruct([2]interface{}{t})
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [2]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]})
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case []string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u})
		}
		v, err := newListStruct(x)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][2]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case *Struct:
		return &Value{Kind: &Value_SingleStruct{SingleStruct: t}}, nil
	case []*Struct:
		v2 := &ListStruct{ListFields: t}
		return &Value{Kind: &Value_ListStruct{ListStruct: v2}}, nil
	default:
	}
	return nil, protoimpl.X.NewError("invalid type: %T", v)
}

// NewStruct constructs a Struct from a generic Go map.
// v is optionally map[string]interface{}.
// The map values are converted using NewValue.
func NewStruct(name string, v ...map[string]interface{}) (*Struct, error) {
	x := &Struct{ClassName: name}
	if v == nil { return x, nil }

	x.Fields = make(map[string]*Value)
	for key, val := range v[0] {
		if !utf8.ValidString(key) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in key: %q", key)
		}
		var err error
		x.Fields[key], err = NewValue(val)
		if err != nil { return nil, err }
	}
	return x, nil
}

func newSingleStruct(v [2]interface{}) (*Struct, error) {
	name, ok := v[0].(string)
	if !ok {
		return nil, protoimpl.X.NewError("need string %T for class name", v[0])
	}
	if v[1] == nil {
		return NewStruct(name)
	}

	return NewStruct(name, v[1].(map[string]interface{}))
}

func newListStruct(v [][2]interface{}) (*ListStruct, error) {
	var err error
	x := make([]*Struct, len(v))
	for i, u := range v {
		x[i], err = newSingleStruct(u)
		if err != nil { return nil, err }
	}
	return &ListStruct{ListFields: x}, nil
}
