package attribute_definition

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToIdentifiableValueDtosKeepsExistingIds(t *testing.T) {
	current := []IdentifiableValue{
		{Id: types.StringValue("id-width"), Value: types.StringValue("Width")},
		{Id: types.StringValue("id-height"), Value: types.StringValue("Height")},
	}
	planned := []IdentifiableValue{
		{Id: types.StringUnknown(), Value: types.StringValue("Height")},
		{Id: types.StringUnknown(), Value: types.StringValue("Depth")},
		{Id: types.StringValue("id-width"), Value: types.StringValue("Width (cm)")},
	}

	result := *ToIdentifiableValueDtos(planned, current)

	if len(result) != 3 {
		t.Fatalf("expected 3 values, got %d", len(result))
	}
	if result[0].Id == nil || *result[0].Id != "id-height" {
		t.Errorf("expected existing ID to be matched on value, got %v", result[0].Id)
	}
	if result[1].Id != nil {
		t.Errorf("expected new value to have no ID, got %s", *result[1].Id)
	}
	if result[2].Id == nil || *result[2].Id != "id-width" || *result[2].Value != "Width (cm)" {
		t.Errorf("expected known ID to be kept for renamed value, got %v", result[2])
	}
}

func TestToIdentifiableValueDtosWithoutState(t *testing.T) {
	planned := []IdentifiableValue{
		{Id: types.StringUnknown(), Value: types.StringValue("Width")},
	}

	result := *ToIdentifiableValueDtos(planned, nil)

	if len(result) != 1 || result[0].Id != nil || *result[0].Value != "Width" {
		t.Errorf("unexpected result %v", result)
	}
}
