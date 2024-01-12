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

func checkItemInBlocks(blocks2 [][]*hclsyntax.Block, labels []string) ([]*hclsyntax.Block, int) {
	for k, blocks := range blocks2 {
		block := blocks[len(blocks)-1]
		var found bool
		if block.Type == labels[0] {
			for i := 1; i < len(labels); i++ {
				if len(block.Labels) > i-1 && block.Labels[i-1] == labels[i] {
					found = true
				} else {
					found = false
					break
				}
			}
		}
		if found {
			return blocks, k
		}
	}
	return nil, -1
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

	var blocks2 [][]*hclsyntax.Block
	for _, item := range body.Blocks {
		labels := append([]string{item.Type}, item.Labels...)
		leading, k := checkItemInBlocks(blocks2, labels)
		if leading != nil {
			blocks2[k] = append(leading, item)
		} else {
			blocks2 = append(blocks2, []*hclsyntax.Block{item})
		}
	}

	for _, blocks := range blocks2 {
		var val interface{}
		var err error
		b0 := blocks[0]
		subnode := node.AddNodes(b0.Type, b0.Labels...)
		if len(blocks) > 1 {
			var multi []map[string]interface{}
			for i, block := range blocks {
				subnode2 := subnode.AddNode(fmt.Sprintf("%d", i))
				x, err := decodeBody(ref, subnode2, file, block.Body)
				if err != nil {
					return nil, err
				}
				multi = append(multi, x)
			}
			val = multi
		} else {
			val, err = decodeBody(ref, subnode, file, b0.Body)
			if err != nil {
				return nil, err
			}
		}

		labels := append([]string{b0.Type}, b0.Labels...)
		var x map[string]interface{}
		for j := len(labels) - 1; j >= 0; j-- {
			if x == nil {
				x = map[string]interface{}{labels[j]: val}
			} else {
				x = map[string]interface{}{labels[j]: x}
			}
		}

		l0 := labels[0]
		if object[l0] == nil {
			object[l0] = x[l0]
		} else {
			loop(x[l0].(map[string]interface{}), object[l0].(map[string]interface{}))
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
