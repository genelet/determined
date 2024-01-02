package dethcl

import (
	"reflect"
	"testing"

	"github.com/genelet/determined/utils"
)

type xclass struct {
	Name    string             `json:"name" hcl:"name"`
	Squares map[string]*square `json:"squares" hcl:"squares"`
	Circles map[string]*circle `json:"circles" hcl:"circles"`
}

func TestMapList(t *testing.T) {
	x := &xclass{Name: "xclass name",
		Squares: map[string]*square{
			"k1": {SX: 1, SY: 2}, "k2": {SX: 3, SY: 4}},
		Circles: map[string]*circle{
			"k5": {5.6}, "k6": {6.7}}}
	bs, err := Marshal(x)
	if err != nil {
		t.Fatal(err)
	}

	typ := reflect.TypeOf(x).Elem()
	n := typ.NumField()
	oriValue := reflect.ValueOf(x).Elem()

	ref := make(map[string]interface{})
	spec := make(map[string]interface{})

	for i := 0; i < n; i++ {
		field := typ.Field(i)
		rawField := oriValue.Field(i)
		if field.Type.Kind() != reflect.Map {
			continue
		}
		var arr []string
		iter := rawField.MapRange()
		for iter.Next() {
			//k := iter.Key()
			v := iter.Value()
			s := v.Type().String()
			arr = append(arr, s)
			ref[s] = reflect.New(v.Type().Elem()).Elem().Addr().Interface()
		}
		spec[field.Name] = arr
	}

	tr, err := utils.NewStruct(typ.Name(), spec)
	xc := &xclass{}
	err = UnmarshalSpec(bs, xc, tr, ref)
	if err != nil {
		t.Fatal(err)
	}

	if x.Squares["k1"].SX != xc.Squares["k1"].SX ||
		x.Squares["k2"].SX != xc.Squares["k2"].SX ||
		x.Circles["k5"].Radius != xc.Circles["k5"].Radius ||
		x.Circles["k6"].Radius != xc.Circles["k6"].Radius {
		t.Errorf("%#v", x.Squares["k1"])
		t.Errorf("%#v", x.Squares["k2"])
		t.Errorf("%#v", x.Circles["k5"])
		t.Errorf("%#v", x.Circles["k6"])
		t.Errorf("%#v", xc.Squares["k1"])
		t.Errorf("%#v", xc.Squares["k2"])
		t.Errorf("%#v", xc.Circles["k5"])
		t.Errorf("%#v", xc.Circles["k6"])
	}
}
