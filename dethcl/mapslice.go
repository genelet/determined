package dethcl

import (
	"fmt"
	"strings"

	"github.com/genelet/determined/utils"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty/gocty"
)

func decodeSlice(bs []byte) ([]interface{}, error) {
	file, diags := hclsyntax.ParseConfig(append([]byte("x = "), bs...), rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}
	tuple, ok := (file.Body.(*hclsyntax.Body).Attributes)["x"].Expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("not resolvable")
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
	str := strings.TrimSpace(string(bs))
	if str[0] == '{' && str[len(str)-1] == '}' {
		return decodeObjectConsExpr(bs)
	}
	file, diags := hclsyntax.ParseConfig(bs, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}

	return decodeBody(file, file.Body.(*hclsyntax.Body))
}

func decodeBody(file *hcl.File, body *hclsyntax.Body) (map[string]interface{}, error) {
	object := make(map[string]interface{})
	for key, item := range body.Attributes {
		val, err := expressionToNative(file, item.Expr)
		if err != nil {
			return nil, err
		}
		object[key] = val
	}

	for _, item := range body.Blocks {
		val, err := decodeBody(file, item.Body)
		if err != nil {
			return nil, err
		}

		labels := append([]string{item.Type}, item.Labels...)
		var x map[string]interface{}
		for j := len(labels) - 1; j >= 0; j-- {
			if x == nil {
				x = map[string]interface{}{labels[j]: val}
			} else {
				x = map[string]interface{}{labels[j]: x}
			}
		}

		if object[item.Type] == nil {
			object[item.Type] = x[item.Type]
		} else {
			loop(x[item.Type].(map[string]interface{}), object[item.Type].(map[string]interface{}))
		}
	}

	return object, nil
}

func loop(x, y map[string]interface{}) {
	for k, v := range x {
		if y[k] == nil {
			y[k] = v
		}
		u, ok := v.(map[string]interface{})
		if ok {
			loop(u, y[k].(map[string]interface{}))
		}
	}
}

func decodeObjectConsExpr(bs []byte) (map[string]interface{}, error) {
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
		return utils.CtyToNative(t.Val)
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
