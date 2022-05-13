package determined

// Struct field's type:
//
// METASingle: a single interface
//
// METASlice: a slice of interfaces
//
// METAMap: a string map for interfaces
//
type METAType int
const (
	METAUnknown METAType = iota
	METASingle
	METASliceSingle
	METAMapSingle
	METASlice
	METAMap
)

// Determined represents struct's interface field
//
// Name: name of struct implmenting the interface
//
// Field: the mapped struct has a DeterminedMap
//
type Determined struct {
	MetaType METAType                 `json:"meta_type,omitempty"`

	SingleName string                 `json:"single_name,omitempty"`
	SingleField DeterminedMap         `json:"single_field,omitempty"`

	SliceName []string                `json:"slice_name,omitempty"`
	SliceField []DeterminedMap        `json:"slice_field,omitempty"`

	MapName map[string]string         `json:"map_name,omitempty"`
	MapField map[string]DeterminedMap `json:"map_field,omitempty"`
}

// DeterminedMap represents field name and its Determined
//
type DeterminedMap map[string]*Determined

func (self *Determined) getPair(dex ...interface{}) (string, DeterminedMap, error) {
	var structName string
	var dmap DeterminedMap
	switch self.MetaType {
	case METAMap:
		key := dex[0].(string)
		if self.MapName != nil {
			structName = self.MapName[key]
		}
		if self.MapField != nil {
			dmap = self.MapField[key]
		}
	case METASlice:
		key := dex[0].(int)
		if self.SliceName != nil && len(self.SliceName) > key {
			structName = self.SliceName[key]
		}
		if self.SliceField != nil && len(self.SliceField) > key {
			dmap = self.SliceField[key]
		}
	default:
		structName = self.SingleName
		dmap = self.SingleField
	}
	return structName, dmap, nil
}
