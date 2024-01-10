package dethcl

import (
	"fmt"
	"log"
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
	log.Printf("decodeMap: %s", string(bs))
	str := strings.TrimSpace(string(bs))
	if str[0] == '{' && str[len(str)-1] == '}' {
		return decodeObjectConsExpr(ref, node, bs)
	}
	file, diags := hclsyntax.ParseConfig(bs, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, (diags.Errs())[0]
	}
	log.Printf("11111: %s", "start")
	return decodeBody(ref, node, file, file.Body.(*hclsyntax.Body))
}

func decodeBody(ref map[string]interface{}, node *utils.Tree, file *hcl.File, body *hclsyntax.Body) (map[string]interface{}, error) {
	object := make(map[string]interface{})
	log.Printf("22222: %s=> %#v", node.Name, body.Attributes)
	for key, item := range body.Attributes {
		val, err := expressionToNative(ref, node, file, key, item.Expr, item)
		if err != nil {
			return nil, err
		}
		object[key] = val
	}
	log.Printf("33333: %#v", node.Name)
	for _, down := range node.Downs {
		log.Printf("44444: %#v", down)
	}

	for i, item := range body.Blocks {
		log.Printf("A55555-%d-%s: %s", i, item.Type, node.Name)
		node.ParentAddNodes(item.Type, item.Labels...)
		subNode := node.GetNode(item.Type, item.Labels...)
		log.Printf("B55555-%d-%s: %s=>%s", i, item.Type, node.Name, subNode.Name)
		val, err := decodeBody(ref, subNode, file, item.Body)
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
	for _, down := range node.Downs {
		log.Printf("99999: %#v", down)
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
		log.Printf("CONS %s=>%#v=>%#v", node.Name, key, item.ValueExpr)
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
		subNode := node.AddNode(fmt.Sprintf("%v", key))
		log.Printf("ARRAY node %s", subNode.Name)
		return decodeSlice(ref, subNode, bs)
	case *hclsyntax.ObjectConsExpr: // map
		rng := t.SrcRange
		bs := file.Bytes[rng.Start.Byte:rng.End.Byte]
		subNode := node.AddNode(fmt.Sprintf("%v", key))
		log.Printf("MAP node %s", subNode.Name)
		return decodeMap(ref, subNode, bs)
	default:
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
}
