# determined

_Determined_ marshals and unmarshals JSON and HCL data to _go struct_ containing interfaces determined at run-time.

[![GoDoc](https://godoc.org/github.com/genelet/determined?status.svg)](https://godoc.org/github.com/genelet/determined)

Please check [this article](https://medium.com/@peterbi_91340/decoding-of-dynamic-json-data-1d4e67318661) for JSON decoding. [Here](https://medium.com/@peterbi_91340/marshal-and-unmarshal-hcl-files-1-3-d7591259a8d6) and [here](https://medium.com/@peterbi_91340/marshal-and-unmarshal-hcl-files-2-2-92cd8af1fe1) for HCL encoding and decoding.

To download, 

```bash
go get github.com/genelet/determined
```

# Convertion among Data Formats HCL, JSON and YAML

## Introduction

Hashicorp Configuration Language ([HCL](https://github.com/hashicorp/hcl)) is a user-friendly data format for structured configuration. It combines parameters and declarative logic in a way that is easily understood by both humans and machines. HCL is integral to Hashicorp’s cloud infrastructure automation tools, such as  `Terraform`  and  `Nomad`. With its robust support for expression syntax, HCL has the potential to serve as a general data format with programming capabilities, making it suitable for use in no-code platforms.

However, in many scenarios, we still need to use popular data formats like JSON and YAML alongside HCL. For instance, Hashicorp products use JSON for data communication via REST APIs, while Docker or Kubernetes management in  `Terraform`  requires YAML.

## Question

An intriguing question arises: Is it possible to convert HCL to JSON or YAML, and vice versa? Could we use HCL as the universal configuration language in projects and generate YAML or JSON with CLI or within  `Terraform`  on the fly?

Unfortunately, the answer is generally no. The expressive power of HCL surpasses that of JSON and YAML. In particular, HCL uses labels to express sorted maps, while JSON and YAML treat all maps as unsorted. Most importantly, HCL allows variables and logic expressions, while JSON and YAML are purely data declarative. Therefore, some features in HCL can never be accurately represented in JSON.

However, in cases where we don’t care about map orders, and there are no variables or logical expressions, but only generic maps, lists, and scalars, then the answer is yes. This type of HCL can be accurately converted to JSON, and vice versa.

> There is a practical advantage of HCL over YAML: HCL is very readable and less prone to errors, while YAML is sensitive to markers like white-space. One can write a configuration in HCL and let a program handle conversion.

Technically, a JSON or YAML string can be unmarshalled into an anonymous map of  _map[string]interface{}_. For seamless conversion, we need to be able to unmarshal any HCL string into an anonymous map, and marshal an anonymous map into a properly formatted HCL string.

## The Package

`determined`  is a GO package to marshal and unmarshal dynamic JSON and HCL contents with interface types. It has a  `convert`  library for conversions among different data formats.

The following functions in  `determined/convert`  can be used for conversion:

-   hcl to json:  _HCLToJSON(raw []byte) ([]byte, error)_
-   hcl to yaml:  _HCLToYAML(raw []byte) ([]byte, error)_
-   json to hcl:  _JSONToHCL(raw []byte) ([]byte, error)_
-   json to yaml:  _JSONToYAML(raw []byte) ([]byte, error)_
-   yaml to hcl:  _YAMLToHCL(raw []byte) ([]byte, error)_
-   yaml to json:  _YAMLToJSON(raw []byte) ([]byte, error)_

If you start with HCL, make sure it contains only primitive data types of maps, lists and scalars.

> In HCL, square brackets are lists and curly brackets are maps. Use  **equal sign  _=_**  and  **comma**  to separate values for  **list** assignment. But no equal sign nor comma for map.

Here is the example to convert HCL to YAML:



    package main  
      
    import (  
        "fmt"  
        "github.com/genelet/determined/convert"  
    )  
      
    func main() {  
        bs := []byte(`parties = [  
      "one",  
      "two",  
      [  
        "three",  
        "four"  
      ],  
      {  
        five = "51"  
        six = 61  
      }  
    ]  
    roads {  
      y = "b"  
      z {  
        za = "aa"  
        zb = 3.14  
      }  
      x = "a"  
      xy = [  
        "ab",  
        true  
      ]  
    }  
    name = "marcus"  
    num = 2  
    radius = 1  
    `)  
        yml, err := convert.HCLToYAML(bs)  
        if err != nil {  
            panic(err)  
        }  
        fmt.Printf("%s\n", yml)  
    }



> Note that HCL is enclosed internally in curly bracket. But the top-level curly bracket should be removed, so it can be accepted by  [the HCL parser](https://pkg.go.dev/github.com/hashicorp/hcl/v2/hclsyntax).

Run the program to get YAML:

    $ go run x.go  
    
    name: marcus  
    num: 2  
    parties:  
        - one  
        - two  
        - - three  
          - four  
        - five: "51"  
          six: 61  
    radius: 1  
    roads:  
        x: a  
        xy:  
            - ab  
            - true  
        "y": b  
        z:  
            za: aa  
            zb: 3.14

## The CLI

In directory  `cmd`, there is a CLI program  `convert.go`. Its usage is

    hcl, json and yaml are choices of the formats  
      
    $ go run convert.go  
    
    convert [options] <filename>  
      -from string  
         from format (default "hcl")  
      -to string  
         to format (default "yaml")
    
    This is a HCL:
    
    version = "3.7"  
    services "db" {  
      image = "hashicorpdemoapp/product-api-db:v0.0.22"  
      ports = [  
        "15432:5432"  
      ]  
      environment {  
        POSTGRES_DB = "products"  
        POSTGRES_USER = "postgres"  
        POSTGRES_PASSWORD = "password"  
      }  
    }  
    services "api" {  
      environment {  
        CONFIG_FILE = "/config/config.json"  
      }  
      depends_on = [  
        "db"  
      ]  
      image = "hashicorpdemoapp/product-api:v0.0.22"  
      ports = [  
        "19090:9090"  
      ]  
      volumes = [  
        "./conf.json:/config/config.json"  
      ]  
    }  

  

Convert it to JSON:

    $ go run convert.go -to json the_above.hcl   
    
    {"services":{"api":{"depends_on":["db"],"environment":{"CONFIG_FILE":"/config/config.json"},"image":"hashicorpdemoapp/product-api:v0.0.22","ports":["19090:9090"],"volumes":["./conf.json:/config/config.json"]},"db":{"environment":{"POSTGRES_DB":"products","POSTGRES_PASSWORD":"password","POSTGRES_USER":"postgres"},"image":"hashicorpdemoapp/product-api-db:v0.0.22","ports":["15432:5432"]}},"version":"3.7"}

Convert it to YAML:

    $ go run convert.go the_above.hcl  
    
    services:  
        api:  
            depends_on:  
                - db  
            environment:  
                CONFIG_FILE: /config/config.json  
            image: hashicorpdemoapp/product-api:v0.0.22  
            ports:  
                - 19090:9090  
            volumes:  
                - ./conf.json:/config/config.json  
        db:  
            environment:  
                POSTGRES_DB: products  
                POSTGRES_PASSWORD: password  
                POSTGRES_USER: postgres  
            image: hashicorpdemoapp/product-api-db:v0.0.22  
            ports:  
                - 15432:5432  
    version: "3.7"

We see that HCL’s syntax is cleaner, more readable, and less error-prone compared to JSON and YAML.

## Summary

HCL is a novel data format that offers advantages over JSON and YAML. In this article, we have demonstrated how to convert data among these three formats.
