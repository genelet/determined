package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/genelet/determined/convert"
)

// treat json as map[string]interface{} or []interface{}
// and convert it into hcl
func main() {
	jsn := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		jsn += scanner.Text()
	}
	if scanner.Err() != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", scanner.Err())
		os.Exit(-1)
	}
	hcl, err := convert.JSONToHCL([]byte(jsn))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("%s\n", hcl)
	os.Exit(0)
}
