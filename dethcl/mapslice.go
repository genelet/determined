package dethcl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func encode(current interface{}, level ...int) ([]byte, error) {
	var str string

	rv := reflect.ValueOf(current)
	switch rv.Kind() {
	case reflect.Map:
		var arr []string
		for name, item := range current.(map[string]interface{}) {
			bs, err := Marshal(item)
			if err != nil {
				return nil, err
			}
			arr = append(arr, fmt.Sprintf("%s = %s", name, bs))
		}
		str = fmt.Sprintf("{\n%s\n}", strings.Join(arr, ",\n"))
	case reflect.Slice, reflect.Array:
		var arr []string
		for _, item := range current.([]interface{}) {
			bs, err := Marshal(item)
			if err != nil {
				return nil, err
			}
			arr = append(arr, string(bs))
		}
		str = fmt.Sprintf("[\n%s\n]", strings.Join(arr, ",\n"))
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
		val, err := getValFromExpression(file, item)
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
		val, err := getValFromExpression(file, item.ValueExpr)
		if err != nil {
			return nil, err
		}
		object[key.AsString()] = val
	}
	return object, nil
}

func getValFromExpression(file *hcl.File, item hclsyntax.Expression) (interface{}, error) {
	switch typ := item.(type) {
	case *hclsyntax.TemplateExpr:
		if typ.IsStringLiteral() {
			val, diags := typ.Value(nil)
			if diags.HasErrors() {
				return nil, (diags.Errs())[0]
			}
			var v string
			err := gocty.FromCtyValue(val, &v)
			return v, err
		} else {
			return nil, fmt.Errorf("template for type %v not implemented", typ)
		}
	case *hclsyntax.LiteralValueExpr:
		val := typ.Val
		switch val.Type() {
		case cty.String:
			var v string
			err := gocty.FromCtyValue(val, &v)
			return v, err
		case cty.Number:
			var v float64
			err := gocty.FromCtyValue(val, &v)
			return v, err
		case cty.Bool:
			var v bool
			err := gocty.FromCtyValue(val, &v)
			return v, err
		default:
			return nil, fmt.Errorf("primitive value %#v not implementned", val)
		}
	case *hclsyntax.TupleConsExpr: // array
		rng := typ.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		return decodeSlice(bs)
	case *hclsyntax.ObjectConsExpr: // map
		rng := typ.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		return decodeMap(bs)
	default:
	}
	return nil, fmt.Errorf("unknow type %T", item)
}
