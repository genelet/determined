package deth

import (
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	utf8 "unicode/utf8"
)

// NewValue constructs a Value from generic Go interface v.
//
//	╔══════════════════╤═══════════════════╤═══════╗
//	║ Go type          │ Conversion        │ Label ║
//	╠══════════════════╪═══════════════════╪═══════╣
//	║ string           │ ending Struct     │  no   ║
//	║ [1]string        │ ending Struct     │  1    ║
//	║ [2]string        │ ending Struct     │  2    ║
//	║ [3]string        │ ending Struct     │  3    ║
//	║ [4]string        │ ending Struct     │  4    ║
//	║ [5]string        │ ending Struct     │  5    ║
//	║ [6]string        │ ending Struct     │  6    ║
//	║ [7]string        │ ending Struct     │  7    ║
//  ║                  │                   │       ║
//	║ [2]interface{}   │ SingleStruct      │  no   ║
//	║ [3]interface{}   │ SingleStruct      │  1    ║
//	║ [4]interface{}   │ SingleStruct      │  2    ║
//	║ [5]interface{}   │ SingleStruct      │  3    ║
//	║ [6]interface{}   │ SingleStruct      │  4    ║
//	║ [7]interface{}   │ SingleStruct      │  5    ║
//	║ [8]interface{}   │ SingleStruct      │  6    ║
//	║ [9]interface{}   │ SingleStruct      │  7    ║
//  ║                  │                   │       ║
//	║ []string         │ ending ListStruct │  no   ║
//	║ [][1]string      │ ending ListStruct │  1    ║
//	║ [][2]string      │ ending ListStruct │  2    ║
//	║ [][3]string      │ ending ListStruct │  3    ║
//	║ [][4]string      │ ending ListStruct │  4    ║
//	║ [][5]string      │ ending ListStruct │  5    ║
//	║ [][6]string      │ ending ListStruct │  6    ║
//	║ [][7]string      │ ending ListStruct │  7    ║
//  ║                  │                   │       ║
//	║ [][2]interface{} │ ListStruct        │  no   ║
//	║ [][3]interface{} │ ListStruct        │  1    ║
//	║ [][4]interface{} │ ListStruct        │  2    ║
//	║ [][5]interface{} │ ListStruct        │  3    ║
//	║ [][6]interface{} │ ListStruct        │  4    ║
//	║ [][7]interface{} │ ListStruct        │  5    ║
//	║ [][8]interface{} │ ListStruct        │  6    ║
//	║ [][9]interface{} │ ListStruct        │  7    ║
//  ║                  │                   │       ║
//  ║ *Struct          │ SingleStruct      │       ║
//  ║ []*Struct        │ ListStruct        │       ║
//	╚══════════════════╧═══════════════════════════╝
func NewValue(v interface{}) (*Value, error) {
	switch t := v.(type) {
	// string is treated as ending Struct without fields
	case string:
		v, err := newSingleStruct([2]interface{}{t}, 0)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [1]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 1)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [2]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 2)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [3]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 3)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [4]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 4)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [5]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 5)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [6]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 6)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [7]string:
		v, err := newSingleStruct([2]interface{}{t[0]}, 7)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil

	case [2]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 0)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [3]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 1)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [4]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 2)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [5]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 3)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [6]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 4)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [7]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 5)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [8]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 6)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [9]interface{}:
		v, err := newSingleStruct([2]interface{}{t[0], t[1]}, 7)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil

	case []string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u})
		}
		v, err := newListStruct(x, 0)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][1]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 1)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][2]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 2)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][3]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 3)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][4]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 4)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][5]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 5)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][6]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 6)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][7]string:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0]})
		}
		v, err := newListStruct(x, 7)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil

	case [][2]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 0)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][3]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 1)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][4]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 2)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][5]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 3)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][6]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 4)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][7]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 5)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][8]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 6)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][9]interface{}:
		var x [][2]interface{}
		for _, u := range t {
			x = append(x, [2]interface{}{u[0], u[1]})
		}
		v, err := newListStruct(x, 7)
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
	return newStruct(name, 0, v...)
}

func newStruct(name string, n int, v ...map[string]interface{}) (*Struct, error) {
	x := &Struct{ClassName: name, NLabels: int32(n)}
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

func newSingleStruct(v [2]interface{}, n int) (*Struct, error) {
	name, ok := v[0].(string)
	if !ok {
		return nil, protoimpl.X.NewError("need string %T for class name", v[0])
	}
	if v[1] == nil {
		return newStruct(name, n)
	}

	return newStruct(name, n, v[1].(map[string]interface{}))
}

func newListStruct(v [][2]interface{}, n int) (*ListStruct, error) {
	var err error
	x := make([]*Struct, len(v))
	for i, u := range v {
		x[i], err = newSingleStruct(u, n)
		if err != nil { return nil, err }
	}
	return &ListStruct{ListFields: x}, nil
}
