package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/genelet/determined/convert"
)

// treat hcl as map[string]interface{} or []interface{}
// and convert it into json
func main() {
	hcl := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		hcl += scanner.Text()
	}
	if scanner.Err() != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", scanner.Err())
		os.Exit(-1)
	}
	jsn, err := convert.HCLToJSON([]byte(hcl))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}
	fmt.Printf("%s\n", jsn)
	os.Exit(0)
}
