package dethcl

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

// clone clones a value via pointer
func clone(old interface{}) interface{} {
	obj := reflect.New(reflect.TypeOf(old).Elem())
	oldVal := reflect.ValueOf(old).Elem()
	newVal := obj.Elem()
	for i := 0; i < oldVal.NumField(); i++ {
		newValField := newVal.Field(i)
		if newValField.CanSet() {
			newValField.Set(oldVal.Field(i))
		}
	}

	return obj.Interface()
}

func tag2(old reflect.StructTag) [2]string {
	for _, tag := range strings.Fields(string(old)) {
		if len(tag) >= 5 && strings.ToLower(tag[:5]) == "hcl:\"" {
			tag = tag[5 : len(tag)-1]
			two := strings.SplitN(tag, ",", 2)
			if len(two) == 2 {
				return [2]string{two[0], two[1]}
			}
			return [2]string{two[0], ""}
		}
	}
	return [2]string{}
}

func hcltag(tag reflect.StructTag) []byte {
	two := tag2(tag)
	return []byte(two[0])
}

func rname() string {
	return fmt.Sprintf("%d.hcl", rand.Int())
}

func ctyToExpression(cv cty.Value, rng hcl.Range) hclsyntax.Expression {
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

func expressionToCty(ref map[string]interface{}, node *Tree, v hclsyntax.Expression) (cty.Value, error) {
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
				return cty.EmptyObjectVal, fmt.Errorf("node %s not found", trv.RootName())
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
				cv, err := expressionToCty(ref, node, p)
				if err != nil {
					return cty.EmptyObjectVal, err
				}
				x, err := ctyToNative(cv)
				if err != nil {
					return cty.EmptyObjectVal, err
				}
				ss = append(ss, x.(string))
			}
			return nativeToCty(strings.Join(ss, ""))
		}
	case *hclsyntax.BinaryOpExpr:
		lcty, err := expressionToCty(ref, node, t.LHS)
		if err != nil {
			return cty.EmptyObjectVal, err
		}
		rcty, err := expressionToCty(ref, node, t.RHS)
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

func nativeToCty(item interface{}) (cty.Value, error) {
	typ, err := gocty.ImpliedType(item)
	if err != nil {
		return cty.EmptyObjectVal, err
	}
	return gocty.ToCtyValue(item, typ)
}

func ctyToNative(val cty.Value) (interface{}, error) {
	ty := val.Type()
	switch ty {
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
	case cty.List(cty.String):
		var v []string
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.List(cty.Number):
		var v []float64
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.List(cty.Bool):
		var v []bool
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.List(cty.DynamicPseudoType):
		var v []interface{}
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.Map(cty.String):
		var v map[string]string
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.Map(cty.Number):
		var v map[string]float64
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.Map(cty.Bool):
		var v map[string]bool
		err := gocty.FromCtyValue(val, &v)
		return v, err
	case cty.Map(cty.DynamicPseudoType):
		var v map[string]interface{}
		err := gocty.FromCtyValue(val, &v)
		return v, err
	default:
	}

	switch {
	case ty.IsListType():
	case ty.IsMapType():
	case ty.IsObjectType(), ty.IsSetType():
		var u map[string]interface{}
		for k, v := range val.AsValueMap() {
			x, err := ctyToNative(v)
			if err != nil {
				return nil, err
			}
			if u == nil {
				u = make(map[string]interface{})
			}
			u[k] = x
		}
		return u, nil
	case ty.IsTupleType():
		var u []interface{}
		for _, v := range val.AsValueSlice() {
			x, err := ctyToNative(v)
			if err != nil {
				return nil, err
			}
			u = append(u, x)
		}
		return u, nil
	case ty.IsPrimitiveType():
	default:
	}

	return nil, fmt.Errorf("assumed primitive value %#v not implementned", val)
}
