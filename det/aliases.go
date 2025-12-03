// Package det provides type aliases for the schema package types.
// This file provides backwards compatibility for code that imports
// types from github.com/genelet/determined/det.
package det

import "github.com/genelet/schema"

// Type aliases for backwards compatibility.
// All these types are now defined in github.com/genelet/schema.
type (
	Struct     = schema.Struct
	Value      = schema.Value
	ListStruct = schema.ListStruct
	MapStruct  = schema.MapStruct
	Map2Struct = schema.Map2Struct

	// Value_SingleStruct and other oneof wrappers
	Value_SingleStruct = schema.Value_SingleStruct
	Value_ListStruct   = schema.Value_ListStruct
	Value_MapStruct    = schema.Value_MapStruct
	Value_Map2Struct   = schema.Value_Map2Struct
)

// NewValue constructs a Value from a generic Go interface.
// See schema.NewValue for full documentation.
var NewValue = schema.NewValue

// NewStruct constructs a Struct specification for dynamic type unmarshaling.
// See schema.NewStruct for full documentation.
var NewStruct = schema.NewStruct
