package dethcl

import (
	"testing"
)

func TestProtobuf(t *testing.T) {
	/*
				data1 := `terraform {
					required_providers {
					  aws = {
						source  = "hashicorp/aws"
						version = "~> 1.0.4"
					  }
					}
				  }

				  variable "aws_region" {}

				  variable "base_cidr_block" {
					description = "A /16 CIDR range definition, such as 10.1.0.0/16, that the VPC will use"
					default = "10.1.0.0/16"
				  }

				  variable "availability_zones" {
					description = "A list of availability zones in which to create subnets"
					type = list(string)
				  }

				  provider "aws" {
					region = var.aws_region
				  }

				  resource "aws_vpc" "main" {
					# Referencing the base_cidr_block variable allows the network address
					# to be changed without modifying the configuration.
					cidr_block = var.base_cidr_block
				  }

				  resource "aws_subnet" "az" {
					# Create one subnet for each given availability zone.
					count = length(var.availability_zones)

					# For each subnet, use one of the specified availability zones.
					availability_zone = var.availability_zones[count.index]

					# By referencing the aws_vpc.main object, Terraform knows that the subnet
					# must be created only after the VPC is created.
					vpc_id = aws_vpc.main.id

					# Built-in functions and operators can be used for simple transformations of
					# values, such as computing a subnet address. Here we create a /20 prefix for
					# each subnet, using consecutive addresses for each availability zone,
					# such as 10.1.16.0/20 .
					cidr_block = cidrsubnet(aws_vpc.main.cidr_block, 4, count.index+1)
				  }`

				err := ParseProtobuf([]byte(data1))
				if err != nil {
					t.Fatal(err)
				}

			data2 := `locals {
				string1       = "str1"
				string2       = "str2"
				int1          = 3
				apply_format  = format("This is %s", local.string1)
				apply_format2 = format("%s_%s_%d", local.string1, local.string2, local.int1)
			   }

			   output "apply_format" {
				value = local.apply_format
			   }
			   output "apply_format2" {
				value = local.apply_format2
			   }`
			err := ParseProtobuf([]byte(data2))
			if err != nil {
				t.Fatal(err)
			}

		data3 := `locals {
			format_list = formatlist("Hello, %s!", ["A", "B", "C"])
		   }

		   output "format_list" {
			value = local.format_list
		   }`
		err := ParseProtobuf([]byte(data3))
		if err != nil {
			t.Fatal(err)
		}

		t.Fatal(err)
	*/
}
