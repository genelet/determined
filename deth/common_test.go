package deth

import (
	"testing"
)

func TestCommonString(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", "label1_1", "label1_2", map[string]interface{}{
			"TheString888": "Circle",
			"TheString777": [2]string{"Circle", "label2_1"},
			"TheString":    [2]interface{}{"Circle"},
			"TheString666": [3]interface{}{"Circle", "label2_2"},

			"TheList888":   []string{"CircleClass1", "CircleClass2"},
			"TheList": [][2]interface{}{
				{"CircleClass1"},
				{"CircleClass2"}},

			"TheHash": [][2]string{
				{"CircleClass1", "a1"},
				{"CircleClass2", "a2"}},
			"TheHash888": [][3]interface{}{
				{"CircleClass1", "a1"},
				{"CircleClass2", "a2"}},
		},
	)
	if err != nil {
		panic(err)
	}
		t.Errorf("%s", endpoint.String())

	fields := endpoint.GetFields()
	if fields["TheString888"].String() != fields["TheString"].String() {
		t.Errorf("%s", fields["TheString888"].String())
		t.Errorf("%s", fields["TheString"].String())
	}
	if fields["TheList888"].String() != fields["TheList"].String() {
		t.Errorf("%s", fields["TheList888"].String())
		t.Errorf("%s", fields["TheList"].String())
	}
	if fields["TheMap888"].String() != fields["TheMap"].String() {
		t.Errorf("%s", fields["TheMap888"].String())
		t.Errorf("%s", fields["TheMap"].String())
	}
}

func TestCommonStruct(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"Shape": [2]interface{}{
				"Class1", map[string]interface{}{"Field1": "Circle"}},
		},
	)
	if err != nil {
		panic(err)
	}
	shapeFields := endpoint.GetFields()
	shapeEndpoint := shapeFields["Shape"].GetSingleStruct()
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field1"].GetSingleStruct()
	if endpoint.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class1" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape endpoint: %s", shapeEndpoint.String())
		t.Errorf("field 1 endpoint: %s", field1Endpoint.String())
	}
}

/*
func TestCommonMap(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"HashShapes": map[string][2]interface{}{
				"x1": [2]interface{}{
					"Class5", map[string]interface{}{"Field4": "Circle"}},
				"y1": [2]interface{}{
					"Class6", map[string]interface{}{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		panic(err)
	}
	shapeFields := endpoint.GetFields()
	shapeEndpoint := shapeFields["HashShapes"].GetListStruct().GetListFields()
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field4"].GetSingleStruct()
	if endpoint.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class5" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape endpoint: %s", shapeEndpoint.String())
		t.Errorf("field 1 endpoint: %s", field1Endpoint.String())
	}
}
*/

func TestCommonList(t *testing.T) {
	endpoint, err := NewStruct(
		"Geo", map[string]interface{}{
			"ListShapes": [][2]interface{}{
				{"Class2", map[string]interface{}{"Field3": "Circle"}},
				{"Class3", map[string]interface{}{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		panic(err)
	}
	shapeFields := endpoint.GetFields()
	shapeEndpoint := shapeFields["ListShapes"].GetListStruct().GetListFields()[1]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field5"].GetSingleStruct()
	if endpoint.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class3" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape endpoint: %s", shapeEndpoint.String())
		t.Errorf("field 1 endpoint: %s", field1Endpoint.String())
	}
}
