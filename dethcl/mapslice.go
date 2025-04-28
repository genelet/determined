package dethcl

import (
	"fmt"
	"strings"

	"github.com/genelet/determined/utils"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func decodeSlice(ref map[string]interface{}, node *utils.Tree, bs []byte) ([]interface{}, error) {
	file, diags := hclsyntax.ParseConfig(append([]byte("x = "), bs...), rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}
	tuple, ok := (file.Body.(*hclsyntax.Body).Attributes)["x"].Expr.(*hclsyntax.TupleConsExpr)
	if !ok {
		return nil, fmt.Errorf("not resolvable")
	}

	var object []interface{}
	for dex, item := range tuple.Exprs {
		val, err := expressionToNative(ref, node, file, dex, item)
		if err != nil {
			return nil, err
		}
		object = append(object, val)
	}
	return object, nil
}

func decodeMap(ref map[string]interface{}, node *utils.Tree, bs []byte) (map[string]interface{}, error) {
	str := strings.TrimSpace(string(bs))
	if str[0] == '{' && str[len(str)-1] == '}' {
		return decodeObjectConsExpr(ref, node, bs)
	}
	file, diags := hclsyntax.ParseConfig(bs, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}

	return decodeBody(ref, node, file, file.Body.(*hclsyntax.Body))
}

func decodeBody(ref map[string]interface{}, node *utils.Tree, file *hcl.File, body *hclsyntax.Body) (map[string]interface{}, error) {
	object := make(map[string]interface{})
	for key, item := range body.Attributes {
		val, err := expressionToNative(ref, node, file, key, item.Expr, item)
		if err != nil {
			return nil, err
		}
		object[key] = val
	}

	var sliceBodies map[string][]*hclsyntax.Body
	counts := make(map[string]int)
	var mapBodies map[string]map[string]*hclsyntax.Body
	var map2Bodies map[string]map[string]map[string]*hclsyntax.Body
	for _, item := range body.Blocks {
		switch len(item.Labels) {
		case 0:
			if sliceBodies == nil {
				sliceBodies = make(map[string][]*hclsyntax.Body)
			}
			sliceBodies[item.Type] = append(sliceBodies[item.Type], item.Body)
			counts[item.Type]++
		case 1:
			if mapBodies == nil {
				mapBodies = make(map[string]map[string]*hclsyntax.Body)
			}
			if mapBodies[item.Type] == nil {
				mapBodies[item.Type] = make(map[string]*hclsyntax.Body)
			}
			mapBodies[item.Type][item.Labels[0]] = item.Body
		case 2:
			if map2Bodies == nil {
				map2Bodies = make(map[string]map[string]map[string]*hclsyntax.Body)
			}
			if map2Bodies[item.Type] == nil {
				map2Bodies[item.Type] = make(map[string]map[string]*hclsyntax.Body)
			}
			if map2Bodies[item.Type][item.Labels[0]] == nil {
				map2Bodies[item.Type][item.Labels[0]] = make(map[string]*hclsyntax.Body)
			}
			map2Bodies[item.Type][item.Labels[0]][item.Labels[1]] = item.Body
		default:
			return nil, fmt.Errorf("not resolvable")
		}
	}

	for key, bodies := range sliceBodies {
		var val []interface{}
		keynode := node.AddNode(key)
		for i, body := range bodies {
			subnode := keynode.AddNode(fmt.Sprintf("%d", i))
			x, err := decodeBody(ref, subnode, file, body)
			if err != nil {
				return nil, err
			}
			val = append(val, x)
		}
		if counts[key] > 1 {
			object[key] = val
		} else {
			object[key] = val[0]
		}
	}

	for key, bodies := range mapBodies {
		val := make(map[string]interface{})
		keynode := node.AddNode(key)
		for k, body := range bodies {
			subnode := keynode.AddNode(k)
			x, err := decodeBody(ref, subnode, file, body)
			if err != nil {
				return nil, err
			}
			val[k] = x
		}
		object[key] = val
	}

	for key, bodies := range map2Bodies {
		val := make(map[string]interface{})
		keynode := node.AddNode(key)
		for k, bodies2 := range bodies {
			val2 := make(map[string]interface{})
			keynode2 := keynode.AddNode(k)
			for k2, body := range bodies2 {
				subnode := keynode2.AddNode(k2)
				x, err := decodeBody(ref, subnode, file, body)
				if err != nil {
					return nil, err
				}
				val2[k2] = x
			}
			val[k] = val2
		}
		object[key] = val
	}

	return object, nil
}

func decodeObjectConsExpr(ref map[string]interface{}, node *utils.Tree, bs []byte) (map[string]interface{}, error) {
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
		val, err := expressionToNative(ref, node, file, key.AsString(), item.ValueExpr)
		if err != nil {
			return nil, err
		}
		object[key.AsString()] = val
	}
	return object, nil
}

func expressionToNative(ref map[string]interface{}, node *utils.Tree, file *hcl.File, key interface{}, item hclsyntax.Expression, attr ...*hclsyntax.Attribute) (interface{}, error) {
	switch t := item.(type) {
	case *hclsyntax.TupleConsExpr: // array
		rng := t.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		subnode := node.AddNode(fmt.Sprintf("%v", key))
		return decodeSlice(ref, subnode, bs)
	case *hclsyntax.ObjectConsExpr: // map
		rng := t.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		subnode := node.AddNode(fmt.Sprintf("%v", key))
		return decodeMap(ref, subnode, bs)
	case *hclsyntax.FunctionCallExpr:
		if t.Name == "null" {
			return nil, nil
		}
	default:
	}

	cv, err := utils.ExpressionToCty(ref, node, item)
	if err != nil {
		return nil, err
	}

	if attr != nil {
		attr[0].Expr = utils.CtyToExpression(cv, attr[0].Expr.Range())
	}
	//item = utils.CtyToExpression(cv, item.Range())

	node.AddItem(fmt.Sprintf("%v", key), cv)

	return utils.CtyToNative(cv)
}
