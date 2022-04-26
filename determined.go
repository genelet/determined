package determined

// Struct field's type:
// METASingle: a single interface
// METASlice: a slice of interfaces
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
// Name: name of struct implmenting the interface
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

func NewSingleDetermined(m METAType, c string, n DeterminedMap) *Determined {
	return &Determined{MetaType:m, SingleName:c, SingleField:n}
}

func NewMapDetermined(c map[string]string, n map[string]DeterminedMap) *Determined {
	return &Determined{MetaType:METAMap, MapName:c, MapField:n}
}

func NewSliceDetermined(c []string, n []DeterminedMap) *Determined {
	return &Determined{MetaType:METASlice, SliceName:c, SliceField:n}
}

func (self *Determined) GetPair(dex ...interface{}) (string, DeterminedMap, error) {
	var structName string
	var dmap DeterminedMap
	switch self.MetaType {
	case METAMap:
		key := dex[0].(string)
		structName = self.MapName[key]
		dmap = self.MapField[key]
	case METASlice:
		key := dex[0].(int)
		structName = self.SliceName[key]
		dmap = self.SliceField[key]
	default:
		structName = self.SingleName
		dmap = self.SingleField
	}
	return structName, dmap, nil
}
