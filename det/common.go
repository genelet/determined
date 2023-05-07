package det

import (
	utf8 "unicode/utf8"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

// NewValue constructs a Value from generic Go interface v.
//
//	╔═══════════════════════════╤═══════════════════════════════════╗
//	║ Go type                   │ Conversion                        ║
//	╠═══════════════════════════╪═══════════════════════════════════╣
//	║ string                    │ ending SingleStruct value         ║
//	║ []string                  │ ending ListStruct value           ║
//	║ map[string]string         │ ending MapStruct value            ║
//	║ [2]interface{}            │ SingleStruct value                ║
//	║ [][2]interface{}          │ ListStruct value                  ║
//	║ map[string][2]interface{} │ MapStruct value                   ║
//	╚═══════════════════════════╧═══════════════════════════════════╝
//
func NewValue(v interface{}) (*Value, error) {
	switch t := v.(type) {
	// string is treated as ending Struct without fields
	case string:
		v2, err := newSingleStruct([2]interface{}{t})
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v2}}, nil
	// []string is treated as ending ListStruct without fields
	case []string:
		output := make([][2]interface{}, len(t))
		for k, s := range t {
			output[k] = [2]interface{}{s}
		}
		return NewValue(output)
	// map[string]string is treated as ending MapStruct without fields
	case map[string]string:
		output := make(map[string][2]interface{})
		for k, s := range t {
			output[k] = [2]interface{}{s}
		}
		return NewValue(output)
	case [2]interface{}:
		v2, err := newSingleStruct(t)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v2}}, nil
	case [][2]interface{}:
		v2, err := newListStruct(t)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_ListStruct{ListStruct: v2}}, nil
	case map[string][2]interface{}:
		v2, err := newMapStruct(t)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_MapStruct{MapStruct: v2}}, nil
	default:
	}
	return nil, protoimpl.X.NewError("invalid type: %T", v)
}

// NewStruct constructs a Struct from a generic Go map.
// The map keys must be valid UTF-8.
// The map values are converted using NewValue.
func NewStruct(name string, v ...map[string]interface{}) (*Struct, error) {
	x := &Struct{ClassName: name}
	if v == nil { return x, nil}
	x.Fields = make(map[string]*Value, len(v[0]))
	for key, val := range v[0] {
		if !utf8.ValidString(key) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", key)
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
		return nil, protoimpl.X.NewError("expect string for the first: %T", v[0])
	}
	x := &Struct{ClassName: name}
	if v[1] == nil {
		return x, nil
	}
	hash, ok := v[1].(map[string]interface{})
	if !ok {
		return nil, protoimpl.X.NewError("expect map[string]interface{} for the second: %T", v[1])
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
