package det

import (
	utf8 "unicode/utf8"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

func stringValue(t string) (*Value, error) {
	v2, err := NewSingleStruct(t)
	if err != nil {
		return nil, err
	}
	return &Value{Kind: &Value_SingleStruct{SingleStruct: v2}}, nil
}

// NewValue constructs a Value from a general-purpose Go interface.
//
//	╔═══════════════════════════╤═════════════════════════════════════════════╗
//	║ Go type                   │ Conversion                                  ║
//	╠═══════════════════════════╪═════════════════════════════════════════════╣
//	║ string                    │ stored as ending SingleStruct value         ║
//	║ []string                  │ stored as ending ListStruct value           ║
//	║ map[string]string         │ stored as ending MapStruct value            ║
//	║ map[string]interface{}    │ stored as SingleStruct value                ║
//	║ [][2]interface{}          │ stored as ListStruct value                  ║
//	║ map[string][2]interface{} │ stored as ListStruct value                  ║
//	╚═══════════════════════════╧═════════════════════════════════════════════╝
//
func NewValue(name string, v interface{}) (*Value, error) {
	switch t := v.(type) {
	// string is treated as ending Struct without fields
	case string:
		return stringValue(t)
	// []string is treated as ending ListStruct without fields
	case []string:
		var output [][2]interface{}
		for _, s := range t {
			output = append(output, [2]interface{}{s})
		}
		v2, err := newListStruct(output)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_ListStruct{ListStruct: v2}}, nil
	// map[string]string is treated as ending MapStruct without fields
	case map[string]string:
		output := make(map[string][2]interface{})
		for k, s := range t {
			output[k] = [2]interface{}{s}
		}
		v2, err := newMapStruct(output)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_MapStruct{MapStruct: v2}}, nil
	case [2]interface{}:
		key, ok := t[0].(string)
		if !ok {
			return nil, protoimpl.X.NewError("invalid string type: %T", t[0])
		}
		if t[1] == nil { return stringValue(key) }
		return NewValue(key, t[1])
	case map[string]interface{}:
		v2, err := NewSingleStruct(name, t)
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
		return nil, protoimpl.X.NewError("invalid type: %T", v)
	}
}

// NewSingleStruct constructs a Struct from a general-purpose Go map.
// The map keys must be valid UTF-8.
// The map values are converted using NewValue.
func NewSingleStruct(name string, v ...map[string]interface{}) (*Struct, error) {
	x := &Struct{ClassName: name}
	if v == nil { return x, nil}
	x.Fields = make(map[string]*Value, len(v[0]))
	for key, val := range v[0] {
		if !utf8.ValidString(key) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", key)
		}
		var err error
		x.Fields[key], err = NewValue(key, val)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

func newListStruct(v [][2]interface{}) (*ListStruct, error) {
	x := &ListStruct{ListFields: make([]*Struct, len(v))}
	for i, u := range v {
		name, ok := u[0].(string)
		if !ok {
			return nil, protoimpl.X.NewError("invalid type second in list: %T. expect string", u[0])
		}
		var err error
		if u[1] == nil {
			x.ListFields[i], err = NewSingleStruct(name)
			if err != nil { return nil, err }
			continue
		}
		hash, ok := u[1].(map[string]interface{})
		if !ok {
			return nil, protoimpl.X.NewError("invalid type third in list: %T. expect map[string]interface{}", u[1])
		}
		x.ListFields[i], err = NewSingleStruct(name, hash)
		if err != nil { return nil, err }
	}
	return x, nil
}

func newMapStruct(v map[string][2]interface{}) (*MapStruct, error) {
	x := &MapStruct{MapFields: make(map[string]*Struct)}
	for i, u := range v {
		name, ok := u[0].(string)
		if !ok {
			return nil, protoimpl.X.NewError("invalid type second in map: %T. expect string", u[0])
		}
		var err error
		if u[1] == nil {
			x.MapFields[i], err = NewSingleStruct(name)
			if err != nil { return nil, err }
			continue
		}
		hash, ok := u[1].(map[string]interface{})
		if !ok {
			return nil, protoimpl.X.NewError("invalid type third in map: %T. expect map[string]interface{}", u[1])
		}
		x.MapFields[i], err = NewSingleStruct(name, hash)
		if err != nil { return nil, err }
	}
	return x, nil
}
