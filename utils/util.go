package utils

import (
	"fmt"
	"math/big"
	"strings"

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

func short(t hcl.Expression, ctx *hcl.EvalContext) (cty.Value, error) {
	cv, diags := t.Value(ctx)
	if diags.HasErrors() {
		return cty.EmptyObjectVal, (diags.Errs())[0]
	}
	return cv, nil
}

func ExpressionToCty(ref map[string]interface{}, node *Tree, v hclsyntax.Expression) (cty.Value, error) {
	if v == nil {
		return cty.NilVal, nil
	}

	switch t := v.(type) {
	case *hclsyntax.FunctionCallExpr:
		if ref[FUNCTIONS] == nil {
			return cty.EmptyObjectVal, fmt.Errorf("function call is nil for %s", t.Name)
		}
		ctx := &hcl.EvalContext{
			Functions: ref[FUNCTIONS].(map[string]function.Function),
			Variables: ref[ATTRIBUTES].(*Tree).Variables(),
		}
		return short(t, ctx)
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
				case hcl.TraverseIndex:
					index, err := CtyNumberToNative(ty.Key)
					if err != nil {
						return cty.EmptyObjectVal, err
					}
					names = append(names, fmt.Sprintf("%v", index))
				default:
				}
			}

			n := len(names)
			name = names[n-1] // the last one is item, not node
			names = names[:n-1]
			n--

			some = ref[ATTRIBUTES].(*Tree)
			if n > 0 && !(n == 1 && names[0] == VAR) {
				some = some.FindNode(names)
				if some == nil {
					return cty.EmptyObjectVal, fmt.Errorf("node not found: %s", trv.RootName())
				}
			}
		}
		return some.Data[name], nil
	case *hclsyntax.TemplateExpr:
		if t.IsStringLiteral() {
			return short(t, nil)
		} else {
			// we may need to enhance the following code to support more complex expressions
			var ss []string
			for _, p := range t.Parts {
				cv, err := ExpressionToCty(ref, node, p)
				if err != nil {
					return cty.EmptyObjectVal, err
				}
				x, err := CtyToNative(cv)
				if err != nil {
					return cty.EmptyObjectVal, err
				}
				if x == nil {
					continue
				}
				ss = append(ss, x.(string))
			}
			if ss == nil {
				return cty.NilVal, nil
			}
			return NativeToCty(strings.Join(ss, ""))
		}
	case *hclsyntax.BinaryOpExpr:
		lcty, err := ExpressionToCty(ref, node, t.LHS)
		if err != nil {
			return cty.EmptyObjectVal, err
		}
		rcty, err := ExpressionToCty(ref, node, t.RHS)
		if err != nil {
			return cty.EmptyObjectVal, err
		}
		return t.Op.Impl.Call([]cty.Value{lcty, rcty})
	case *hclsyntax.ForExpr: // to be implemented
	case *hclsyntax.IndexExpr: // to be implemented
	case *hclsyntax.ParenthesesExpr: // to be implemented and so on...
	default:
	}

	return short(v, nil)
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
		return x, err
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
