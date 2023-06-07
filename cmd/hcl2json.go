package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/genelet/determined/dethcl"
	"os"
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
	obj := map[string]interface{}{}
	if err := dethcl.Unmarshal([]byte(hcl), &obj); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}

	json, err := json.Marshal(obj)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}

	fmt.Printf("%s\n", json)
	os.Exit(0)
}
