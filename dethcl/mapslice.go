package dethcl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty/gocty"
)

func isHashAll(item interface{}) (map[string]interface{}, bool) {
	var foundItem bool
	next, ok := item.(map[string]interface{})
	if ok {
		for _, v := range next {
			_, ok1 := v.(map[string]interface{})
			if ok1 {
				foundItem = true
			} else {
				foundItem = false
			}
		}
	}
	return next, foundItem
}

func encode(current interface{}, level int) ([]byte, error) {
	var str string

	leading := strings.Repeat("  ", level+1)
	lessLeading := strings.Repeat("  ", level)

	rv := reflect.ValueOf(current)
	switch rv.Kind() {
	case reflect.Map:
		var arr []string
		for name, item := range current.(map[string]interface{}) {
			next, foundItem := isHashAll(item)
			if foundItem {
				for key, value := range next {
					next2, found3 := isHashAll(value)
					if found3 {
						for k, v := range next2 {
							name2 := name + ` "` + key + `" "` + k + `"`
							bs, err := mmarshal(v, level+1)
							if err != nil {
								return nil, err
							}
							arr = append(arr, fmt.Sprintf("%s %s", name2, bs))
						}
					} else {
						name2 := name + ` "` + key + `"`
						bs, err := mmarshal(value, level+1)
						if err != nil {
							return nil, err
						}
						arr = append(arr, fmt.Sprintf("%s %s", name2, bs))
					}
				}
			} else {
				bs, err := mmarshal(item, level+1)
				if err != nil {
					return nil, err
				}
				arr = append(arr, fmt.Sprintf("%s = %s", name, bs))
			}
		}
		if level == 0 {
			str = fmt.Sprintf("\n%s\n%s", leading+strings.Join(arr, "\n"+leading), lessLeading)
		} else {
			str = fmt.Sprintf("{\n%s\n%s}", leading+strings.Join(arr, "\n"+leading), lessLeading)
		}
		/*
			var found bool
			for {
				var inner_found bool
				switch t := item.(type) {
				case map[string]interface{}:
					if len(t) == 1 {
						for k, v := range t {
							name += " " + k
							item = v
						}
						inner_found = true
					}
				default:
				}
				if !inner_found {
					break
				}
				found = inner_found
			}
			bs, err := mmarshal(item, level+1)
			if err != nil {
				return nil, err
			}
			if found {
				arr = append(arr, fmt.Sprintf("%s = %s", name, bs))
			} else {
				arr = append(arr, fmt.Sprintf("%s = %s", name, bs))
			}
		*/
	case reflect.Slice, reflect.Array:
		var arr []string
		for _, item := range current.([]interface{}) {
			bs, err := mmarshal(item, level+1)
			if err != nil {
				return nil, err
			}
			arr = append(arr, string(bs))
		}
		str = fmt.Sprintf("[\n%s\n%s]", leading+strings.Join(arr, ",\n"+leading), lessLeading)
	case reflect.String:
		str = "\"" + rv.String() + "\""
	case reflect.Bool:
		str = fmt.Sprintf("%t", rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str = fmt.Sprintf("%d", rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		str = fmt.Sprintf("%d", rv.Uint())
	case reflect.Float32, reflect.Float64:
		str = fmt.Sprintf("%f", rv.Float())
	default:
		return nil, fmt.Errorf("data type %v not supported", rv.Kind())
	}

	return []byte(str), nil
}

func decodeSlice(bs []byte) ([]interface{}, error) {
	file, diags := hclsyntax.ParseConfig(append([]byte("xx = "), bs...), rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}
	tuple, ok := (file.Body.(*hclsyntax.Body).Attributes)["xx"].Expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("not resolvale")
	}

	var object []interface{}
	for _, item := range tuple.Exprs {
		val, err := expressionToNative(file, item)
		if err != nil {
			return nil, err
		}
		object = append(object, val)
	}
	return object, nil
}

func decodeMap(bs []byte) (map[string]interface{}, error) {
	file, diags := hclsyntax.ParseConfig(append([]byte("x = "), bs...), rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}
	exprs, ok := (file.Body.(*hclsyntax.Body).Attributes)["x"].Expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return nil, fmt.Errorf("not resolvale")
	}

	object := make(map[string]interface{})
	for _, item := range exprs.Items {
		key, diags := item.KeyExpr.(*hclsyntax.ObjectConsKeyExpr).Value(nil)
		if diags.HasErrors() {
			return nil, (diags.Errs())[0]
		}
		val, err := expressionToNative(file, item.ValueExpr)
		if err != nil {
			return nil, err
		}
		object[key.AsString()] = val
	}
	return object, nil
}

func expressionToNative(file *hcl.File, item hclsyntax.Expression) (interface{}, error) {
	switch t := item.(type) {
	case *hclsyntax.TemplateExpr:
		if t.IsStringLiteral() {
			val, diags := t.Value(nil)
			if diags.HasErrors() {
				return nil, (diags.Errs())[0]
			}
			var v string
			err := gocty.FromCtyValue(val, &v)
			return v, err
		} else {
			return nil, fmt.Errorf("template for type %v not implemented", t)
		}
	case *hclsyntax.LiteralValueExpr:
		return ctyToNative(t.Val)
	case *hclsyntax.TupleConsExpr: // array
		rng := t.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		return decodeSlice(bs)
	case *hclsyntax.ObjectConsExpr: // map
		rng := t.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		return decodeMap(bs)
	default:
	}
	return nil, fmt.Errorf("unknow type %T", item)
}

func decodeMapMore(node *Tree, file *hcl.File, bd *hclsyntax.Body, current interface{}, labels ...string) error {
	hash := current.(*map[string]interface{})
	for k, v := range bd.Attributes {
		u, err := expressionToNative(file, v.Expr)
		if err != nil {
			return err
		}
		(*hash)[k] = u
	}
	for _, block := range bd.Blocks {
		k := block.Type
		_, ok := (*hash)[k]
		if !ok {
			(*hash)[k] = make(map[string]interface{})
		}
		u := (*hash)[k].(map[string]interface{})
		for _, v := range block.Labels {
			if u[v] == nil {
				u[v] = make(map[string]interface{})
			}
			u = u[v].(map[string]interface{})
		}
		bs := file.Bytes[block.OpenBraceRange.Start.Byte+1 : block.CloseBraceRange.End.Byte-1]
		item := map[string]interface{}{}
		err := unmarshalSpec(node, bs, &item, nil, nil, block.Labels...)
		if err != nil {
			return err
		}
		for k, v := range item {
			u[k] = v
		}
	}
	return nil
}

func decodeSliceMore(node *Tree, file *hcl.File, bd *hclsyntax.Body, current interface{}, labels ...string) error {
	slice := current.(*[]interface{})
	for _, v := range bd.Attributes {
		u, err := expressionToNative(file, v.Expr)
		if err != nil {
			return err
		}
		*slice = append(*slice, u)
	}
	for _, block := range bd.Blocks {
		bs := file.Bytes[block.OpenBraceRange.Start.Byte+1 : block.CloseBraceRange.End.Byte-1]
		next := make(map[string]interface{})
		err := unmarshalSpec(node, bs, &next, nil, nil, block.Labels...)
		if err != nil {
			return err
		}
		*slice = append(*slice, next)
	}
	return nil
}
