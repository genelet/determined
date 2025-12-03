package det

import (
	"testing"
)

func TestCommonString(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"TheString888": "Circle",
			"TheString":    [2]any{"Circle"},
			"TheList888":   []string{"CircleClass1", "CircleClass2"},
			"TheList":      [][2]any{{"CircleClass1"}, {"CircleClass2"}},
			"TheMap888":    map[string]string{"a1": "CircleClass1", "a2": "CircleClass2"},
			"TheMap":       map[string][2]any{"a1": {"CircleClass1"}, "a2": {"CircleClass2"}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	fields := spec.GetFields()
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
	spec, err := NewStruct(
		"Geo", map[string]any{
			"Shape1": [2]any{
				"Class1", map[string]any{"Field1": "Circle1"}},
			"Shape2": [2]any{
				"Class2", map[string]any{"Field2": []string{"Circle2", "Circle3"}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	shapeFields := spec.GetFields()

	shapeEndpoint := shapeFields["Shape1"].GetSingleStruct()
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field1"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class1" ||
		field1Endpoint.ClassName != "Circle1" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}

	shape2Endpoint := shapeFields["Shape2"].GetSingleStruct()
	field2Fields := shape2Endpoint.GetFields()
	field2Endpoint := field2Fields["Field2"].GetListStruct()
	if spec.ClassName != "Geo" ||
		shape2Endpoint.ClassName != "Class2" ||
		field2Endpoint.ListFields[0].ClassName != "Circle2" ||
		field2Endpoint.ListFields[1].ClassName != "Circle3" {
		t.Errorf("shape spec: %s", shape2Endpoint.String())
		t.Errorf("field 2 spec: %s", field2Endpoint.String())
	}
}

func TestCommonList(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"ListShapes": [][2]any{
				{"Class2", map[string]any{"Field3": "Circle"}},
				{"Class3", map[string]any{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	shapeFields := spec.GetFields()
	shapeEndpoint := shapeFields["ListShapes"].GetListStruct().GetListFields()[1]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field5"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class3" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}
}

func TestCommonMap(t *testing.T) {
	spec, err := NewStruct(
		"Geo", map[string]any{
			"HashShapes": map[string][2]any{
				"x1": {"Class5", map[string]any{"Field4": "Circle"}},
				"y1": {"Class6", map[string]any{"Field5": "Circle"}}},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	shapeFields := spec.GetFields()
	shapeEndpoint := shapeFields["HashShapes"].GetMapStruct().GetMapFields()["x1"]
	field1Fields := shapeEndpoint.GetFields()
	field1Endpoint := field1Fields["Field4"].GetSingleStruct()
	if spec.ClassName != "Geo" ||
		shapeEndpoint.ClassName != "Class5" ||
		field1Endpoint.ClassName != "Circle" {
		t.Errorf("shape spec: %s", shapeEndpoint.String())
		t.Errorf("field 1 spec: %s", field1Endpoint.String())
	}
}
