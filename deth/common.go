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
//	║ [2]string        │ ending Struct     │  1    ║
//	║ [3]string        │ ending Struct     │  2    ║
//	║ [4]string        │ ending Struct     │  3    ║
//	║ [5]string        │ ending Struct     │  4    ║
//	║ [6]string        │ ending Struct     │  5    ║
//	║ [7]string        │ ending Struct     │  6    ║
//	║ [8]string        │ ending Struct     │  7    ║
//	║                  │                   │       ║
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
//	║ [][2]string      │ ending ListStruct │  1    ║
//	║ [][3]string      │ ending ListStruct │  2    ║
//	║ [][4]string      │ ending ListStruct │  3    ║
//	║ [][5]string      │ ending ListStruct │  4    ║
//	║ [][6]string      │ ending ListStruct │  5    ║
//	║ [][7]string      │ ending ListStruct │  6    ║
//	║ [][8]string      │ ending ListStruct │  7    ║
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
		v, err := newSingleStruct([]interface{}{t}, 2)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [2]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 2)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [2]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 3)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [3]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 3)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [3]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 4)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [4]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 4)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [4]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 5)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [5]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 5)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [5]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 6)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [6]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 6)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [6]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 7)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [7]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 7)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [7]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 8)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [8]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 8)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [8]string:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 9)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil
	case [9]interface{}:
		var x []interface{}
		for _, s := range t { x = append(x, s) }
		v, err := newSingleStruct(x, 9)
		if err != nil { return nil, err }
		return &Value{Kind: &Value_SingleStruct{SingleStruct: v}}, nil

	case []string:
		var x [][]interface{}
		for _, u := range t {
			y := []interface{}{u}
			x = append(x, y)
		}
		v, err := newListStruct(x, 2)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][2]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 2)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][2]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 3)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][3]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 3)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][3]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 4)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][4]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 4)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][4]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 5)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][5]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 5)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][5]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 6)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][6]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 6)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][6]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 7)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][7]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 7)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][7]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 8)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][8]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 8)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][8]string:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 9)
		if err != nil { return nil, err }
        return &Value{Kind: &Value_ListStruct{ListStruct: v}}, nil
	case [][9]interface{}:
		var x [][]interface{}
		for _, u := range t {
			var y []interface{}
			for _, s := range u { y = append(y, s) }
			x = append(x, y)
		}
		v, err := newListStruct(x, 9)
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
// v... should be string labels, up to 7.
// Plus optionally map[string]interface{}.
// The map values are converted using NewValue.
func NewStruct(name string, v ...interface{}) (*Struct, error) {
	x := &Struct{ClassName: name}
	if v == nil { return x, nil }

TOP:
	for i, value := range v {
		switch t := value.(type) {
		case string:
			switch i {
			case 0: x.Label1 = t
			case 1: x.Label2 = t
			case 2: x.Label3 = t
			case 3: x.Label4 = t
			case 4: x.Label5 = t
			case 5: x.Label6 = t
			case 6: x.Label7 = t
			default:
			}
		case map[string]interface{}:
			x.Fields = make(map[string]*Value, len(t))
			for key, val := range t {
				if !utf8.ValidString(key) {
					return nil, protoimpl.X.NewError("invalid UTF-8 in key: %q", key)
				}
				var err error
				x.Fields[key], err = NewValue(val)
				if err != nil { return nil, err }
			}
			break TOP
		default:
		}
	}
	return x, nil
}

func newSingleStruct(v []interface{}, n int) (*Struct, error) {
	m := len(v)
	if (n!=m && (n-1)!=m) || m<1 {
		return nil, protoimpl.X.NewError("wrong counts: m %d => n %d", m, n)
	}

	name, ok := v[0].(string)
	if !ok {
		return nil, protoimpl.X.NewError("need string %T for class name", v[0])
	}

	return NewStruct(name, v[1:]...)
}

func newListStruct(v [][]interface{}, n int) (*ListStruct, error) {
	var err error
	x := make([]*Struct, len(v))
	for i, u := range v {
		x[i], err = newSingleStruct(u, n)
		if err != nil { return nil, err }
	}
	return &ListStruct{ListFields: x}, nil
}
