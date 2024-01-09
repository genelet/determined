package main

import (
	"fmt"
	"os"
	"github.com/genelet/determined/convert"
)

func main() {
	if len(os.Args) <= 1 {
		os.Exit(-1)
	}
	filename := os.Args[1]

		fmt.Printf("\n\nFILE: %s\n", filename)
		bs, err := os.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		jsn, err := convert.HCLToJSON(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\b", jsn)

	os.Exit(0)
}
