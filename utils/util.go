package utils

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

func CtyToExpression(cv cty.Value, rng hcl.Range) hclsyntax.Expression {
	switch cv.Type() {
	case cty.String, cty.Number, cty.Bool:
		return &hclsyntax.LiteralValueExpr{Val: cv, SrcRange: rng}
	case cty.List(cty.String), cty.List(cty.Number), cty.List(cty.Bool):
		var exprs []hclsyntax.Expression
		for _, item := range cv.AsValueSlice() {
			exprs = append(exprs, &hclsyntax.LiteralValueExpr{Val: item, SrcRange: rng})
		}
		return &hclsyntax.TupleConsExpr{Exprs: exprs, SrcRange: rng}
	case cty.Map(cty.String), cty.Map(cty.Number), cty.Map(cty.Bool):
		var items []hclsyntax.ObjectConsItem
		for k, item := range cv.AsValueMap() {
			items = append(items, hclsyntax.ObjectConsItem{
				KeyExpr:   &hclsyntax.LiteralValueExpr{Val: cty.StringVal(k), SrcRange: rng},
				ValueExpr: &hclsyntax.LiteralValueExpr{Val: item, SrcRange: rng},
			})
		}
		return &hclsyntax.ObjectConsExpr{Items: items, SrcRange: rng}
	default:
	}
	// just use the default seems to be ok
	return &hclsyntax.LiteralValueExpr{Val: cv, SrcRange: rng}
}

func callToCty(ref map[string]interface{}, node *Tree, funcs map[string]interface{}, u *hclsyntax.FunctionCallExpr) (cty.Value, error) {
	if u.Name == "" {
		return cty.EmptyObjectVal, fmt.Errorf("function call is empty")
	}
	if funcs == nil {
		return cty.EmptyObjectVal, fmt.Errorf("function call is nil for %s", u.Name)
	}
	fn, ok := funcs[u.Name]
	if !ok {
		return cty.EmptyObjectVal, fmt.Errorf("function call is not found for %s", u.Name)
	}
	k := -1
	f := reflect.ValueOf(fn)

	n := 0
	var in []reflect.Value
	for _, arg := range u.Args {
		cv, err := ExpressionToCty(ref, node, arg)
		if err != nil {
			return cty.EmptyObjectVal, err
		}
		v, err := CtyToNative(cv)
		if err != nil {
			return cty.EmptyObjectVal, err
		}
		in = append(in, reflect.ValueOf(v))
		n++
	}
	if f.Type().NumIn() != n {
		return cty.EmptyObjectVal, fmt.Errorf("function %s needs %d args, got %d", u.Name, f.Type().NumIn(), n)
	}

	outputs := f.Call(in)
	m := len(outputs)
	for i, output := range outputs {
		if output.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			k = i
			break
		}
	}

	if k >= 0 {
		if k >= m {
			return cty.EmptyObjectVal, fmt.Errorf("error output is missing or mismatched")
		}
		if outputs[k].IsValid() && !outputs[k].IsNil() {
			return cty.EmptyObjectVal, outputs[k].Interface().(error)
		}
		if m == 1 {
			return cty.EmptyObjectVal, nil
		}
	}

	for i, output := range outputs {
		if i == k {
			continue
		}
		// in this version, we assume only one item returned
		return NativeToCty(output.Interface())
	}

	return cty.EmptyObjectVal, nil
}

func ExpressionToCty(ref map[string]interface{}, node *Tree, v hclsyntax.Expression) (cty.Value, error) {
	if v == nil {
		return cty.NilVal, nil
	}

	ctx := new(hcl.EvalContext)
	if ref != nil && ref[ATTRIBUTES] != nil {
		ctx.Variables = ref[ATTRIBUTES].(*Tree).Variables()
	}

	if ref != nil && ref[FUNCTIONS] != nil {
		if u, ok := v.(*hclsyntax.FunctionCallExpr); ok {
			if ref[FUNCTIONS] == nil {
				return cty.EmptyObjectVal, fmt.Errorf("function call is nil for %s", u.Name)
			}
			switch t := ref[FUNCTIONS].(type) {
			case map[string]function.Function:
				ctx.Functions = t
			case map[string]interface{}:
				return callToCty(ref, node, t, u)
			default:
				return cty.EmptyObjectVal, fmt.Errorf("function call is not a map for %s", u.Name)
			}
		}
	}

	cv, diags := v.Value(ctx)
	if diags.HasErrors() {
		return cty.EmptyObjectVal, (diags.Errs())[0]
	}
	return cv, nil
}

func NativeToCty(item interface{}) (cty.Value, error) {
	typ, err := gocty.ImpliedType(item)
	if err != nil {
		return cty.EmptyObjectVal, err
	}
	return gocty.ToCtyValue(item, typ)
}

func CtyNumberToNative(val cty.Value) (interface{}, error) {
	v := val.AsBigFloat()
	if _, accuracy := v.Int64(); accuracy == big.Exact || accuracy == big.Above {
		var x int64
		err := gocty.FromCtyValue(val, &x)
		if x > 0x7FFFFFFF || x < -0x80000000 {
			return x, err
		}
		return int(x), err
	} else if _, accuracy := v.Int(nil); accuracy == big.Exact || accuracy == big.Above {
		var x int
		err := gocty.FromCtyValue(val, &x)
		return x, err
	} else if _, accuracy := v.Float32(); accuracy == big.Exact || accuracy == big.Above {
		var x float32
		err := gocty.FromCtyValue(val, &x)
		return x, err
	}
	var x float64
	err := gocty.FromCtyValue(val, &x)
	return x, err
}

func CtyToNative(val cty.Value) (interface{}, error) {
	if val.IsNull() {
		return nil, nil
	}

	ty := val.Type()
	switch ty {
	case cty.String:
		var v string
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.Number:
		return CtyNumberToNative(val)
	case cty.Bool:
		var v bool
		err := gocty.FromCtyValue(val, &v)
		return v, err
	default:
	}

	switch {
	case ty.IsObjectType(), ty.IsMapType():
		var u map[string]interface{}
		for k, v := range val.AsValueMap() {
			x, err := CtyToNative(v)
			if err != nil {
				return nil, err
			}
			if x == nil {
				continue
			}
			if u == nil {
				u = make(map[string]interface{})
			}
			u[k] = x
		}
		return u, nil
	case ty.IsListType(), ty.IsTupleType(), ty.IsSetType():
		var u []interface{}
		for _, v := range val.AsValueSlice() {
			x, err := CtyToNative(v)
			if err != nil {
				return nil, err
			}
			if x == nil {
				continue
			}
			u = append(u, x)
		}
		return u, nil
	default:
	}

	return nil, fmt.Errorf("assumed primitive value %#v not implementned", val)
}
