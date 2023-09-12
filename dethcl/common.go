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
//	║                  │                   ║
//	║ [2]string        │ ending Ctx        ║
//	║ [][2]string      │ ending Ctx        ║
//	║                  │                   ║
//	║ []string         │ ending ListStruct ║
//	║ [][2]interface{} │ ListStruct        ║
//	║                  │                   ║
//	║ *Struct          │ SingleStruct      ║
//	║ []*Struct        │ ListStruct        ║
//	╚══════════════════╧═══════════════════╝
func NewValue(v interface{}) (*Value, error) {
	switch t := v.(type) {
	// string is treated as ending Struct without fields
	case string:
		v, err := newSingleStruct([2]interface{}{t})
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case []string:
		output := make([][2]interface{}, len(t))
		for k, s := range t {
			output[k] = [2]interface{}{s}
		}
		return NewValue(output)

	// string[2] is treated as ending Struct without fields but service
	case [2]string:
		v2, err := newSingleStruct([2]interface{}{t[0], t[1]})
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v2}}, nil
	// [][2]string is treated as ending ListStruct without fields but service
	case [][2]string:
		output := make([][2]interface{}, len(t))
		for k, s := range t {
			output[k] = [2]interface{}{s[0], s[1]}
		}
		return NewValue(output)

	case [2]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]})
		if err != nil {
			return nil, err
		}
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [][2]interface{}:
		v, err := newListStruct(t)
		if err != nil {
			return nil, err
		}
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
func NewStruct(name string, v ...interface{}) (*Struct, error) {
	x := &Struct{ClassName: name}
	if v == nil {
		return x, nil
	}

	switch t := v[0].(type) {
	case string:
		x.ServiceName = t
	case map[string]interface{}:
		x.Fields = make(map[string]*Value, len(t))
		for key, val := range t {
			if !utf8.ValidString(key) {
				return nil, protoimpl.X.NewError("invalid UTF8 in key: %q", key)
			}
			var err error
			x.Fields[key], err = NewValue(val)
			if err != nil {
				return nil, err
			}
		}
	default:
	}

	return x, nil
}

func newSingleStruct(v [2]interface{}) (*Struct, error) {
	name, ok := v[0].(string)
	if !ok {
		return nil, protoimpl.X.NewError("need string %T for class name", v[0])
	}
	return NewStruct(name, v[1:]...)
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
