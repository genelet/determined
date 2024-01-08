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
	dir := os.Args[1]
	if dir == "" {
		panic("no directory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		filename := entry.Name()
		fmt.Printf("\n\nFILE: %s\n", filename)
		bs, err := os.ReadFile(filename + "/main.tf")
		if err != nil {
			panic(err)
		}
		jsn, err := convert.HCLToJSON(bs)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\b", jsn)
	}
	os.Exit(0)	
}
