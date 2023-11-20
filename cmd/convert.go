package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/genelet/determined/utils"
)

var from string
var to string

func init() {
	flag.StringVar(&from, "from", "hcl", "from format")
	flag.StringVar(&to, "to", "yaml", "to format")
	flag.Parse()
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [options] <filename>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(-1)
}

func main() {
	if from == to {
		fmt.Fprintf(os.Stderr, "error: from and to format are the same\n")
		os.Exit(-1)
	}

	filename := flag.Arg(0)
	if filename == "" {
		usage()
	}

	raw, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}

	switch from {
	case "json":
		switch to {
		case "yaml":
			raw, err = utils.JSONToYAML(raw)
		case "hcl":
			raw, err = utils.JSONToHCL(raw)
		default:
			fmt.Fprintf(os.Stderr, "error: unsupported to format %s\n", to)
			os.Exit(-1)
		}
	case "yaml":
		switch to {
		case "json":
			raw, err = utils.YAMLToJSON(raw)
		case "hcl":
			raw, err = utils.YAMLToHCL(raw)
		default:
			fmt.Fprintf(os.Stderr, "error: unsupported to format %s\n", to)
			os.Exit(-1)
		}
	case "hcl":
		switch to {
		case "json":
			raw, err = utils.HCLToJSON(raw)
		case "yaml":
			raw, err = utils.HCLToYAML(raw)
		default:
			fmt.Fprintf(os.Stderr, "error: unsupported to format %s\n", to)
			os.Exit(-1)
		}
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported from format %s\n", from)
		os.Exit(-1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(-1)
	}

	fmt.Printf("%s\n", raw)
	os.Exit(0)
}
