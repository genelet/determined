package dethcl

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/k0kubun/pp/v3"
)

func ParseProtobuf(dat []byte) error {
	file, diags := hclsyntax.ParseConfig(dat, rname(), hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return diags
	}
	body := file.Body.(*hclsyntax.Body)
	attrs := body.Attributes
	blocks := body.Blocks

	pp.Println(attrs)
	pp.Println(blocks)
	return nil
}
