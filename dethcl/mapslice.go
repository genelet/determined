package dethcl

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
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

func expressionToCty(ref map[string]interface{}, node *Tree, v hclsyntax.Expression) (*cty.Value, error) {
	switch t := v.(type) {
	case *hclsyntax.FunctionCallExpr:
		if ref[FUNCTIONS] == nil {
			return nil, fmt.Errorf("function call is nil for %s", t.Name)
		}
		ctx := &hcl.EvalContext{
			Functions: ref[FUNCTIONS].(map[string]function.Function),
			Variables: ref[ATTRIBUTES].(*Tree).Variables(),
		}
		cv, err := t.Value(ctx)
		if err != nil {
			return nil, err
		}
		return &cv, nil
	case *hclsyntax.ScopeTraversalExpr:
		trv := t.AsTraversal()
		some := node
		name := trv.RootName()
		if !trv.IsRelative() {
			var names []string
			for _, item := range trv {
				switch ty := item.(type) {
				case hcl.TraverseRoot:
					names = append(names, ty.Name)
				case hcl.TraverseAttr:
					names = append(names, ty.Name)
				default:
				}
			}
			n := len(names)
			// the last one is not node but item
			name = names[n-1]
			names = names[:n-1]
			n--

			top := ref[ATTRIBUTES].(*Tree)
			// check the first one
			if n != 0 && names[0] == VAR {
				names = names[1:]
				n--
			}
			some = top.FindNode(names)
			if some == nil {
				return nil, fmt.Errorf("node %s not found", trv.RootName())
			}
		}
		return some.Data[name], nil
	case *hclsyntax.TemplateExpr:
		if t.IsStringLiteral() {
			cv, diags := t.Value(nil)
			if diags.HasErrors() {
				return nil, (diags.Errs())[0]
			}
			return &cv, nil
		} else {
			// multiple expressions as Parts
			var ss []string
			for _, p := range t.Parts {
				c, err := expressionToCty(ref, node, p)
				if err != nil {
					return nil, err
				}
				var cv cty.Value
				if c == nil {
					cv = p.(*hclsyntax.LiteralValueExpr).Val
				} else {
					cv = *c
				}
				x, err := ctyToNative(cv)
				if err != nil {
					return nil, err
				}
				ss = append(ss, x.(string))
			}
			cv, err := nativeToCty(strings.Join(ss, ""))
			if err != nil {
				return nil, err
			}
			return &cv, nil
		}
	case *hclsyntax.BinaryOpExpr:
		lcty, err := expressionToCty(ref, node, t.LHS)
		if err != nil {
			return nil, err
		}
		rcty, err := expressionToCty(ref, node, t.RHS)
		if err != nil {
			return nil, err
		}
		cv, err := t.Op.Impl.Call([]cty.Value{*lcty, *rcty})
		if err != nil {
			return nil, err
		}
		return &cv, err
	default:
	}

	cv, err := v.Value(nil)
	if err != nil {
		return nil, err
	}
	return &cv, nil
}

func nativeToCty(item interface{}) (cty.Value, error) {
	typ, err := gocty.ImpliedType(item)
	if err != nil {
		return cty.EmptyObjectVal, err
	}
	return gocty.ToCtyValue(item, typ)
}

func ctyToNative(val cty.Value) (interface{}, error) {
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
	}
	return nil, fmt.Errorf("primitive value %#v not implementned", val)
}
