package attribute_definition

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var enumValueType = types.ObjectType{AttrTypes: map[string]attr.Type{
	"metadata": types.StringType,
	"number":   types.StringType,
	"value":    types.StringType,
	"value_id": types.StringType,
}}

var null = types.StringNull()

func s(v string) types.String { return types.StringValue(v) }

func enumValue(value string, valueId, number types.String) EnumValue {
	return EnumValue{
		Metadata: types.StringNull(),
		Number:   number,
		Value:    types.StringValue(value),
		ValueId:  valueId,
	}
}

func enumList(t *testing.T, values ...EnumValue) types.List {
	t.Helper()
	l, diags := types.ListValueFrom(context.Background(), enumValueType, values)
	if diags.HasError() {
		t.Fatal(diags)
	}
	return l
}

// modify runs the plan modifier. Like Terraform, the plan starts as the config
// with the computed attributes copied from the state by index.
func modify(t *testing.T, state types.List, config ...EnumValue) []EnumValue {
	t.Helper()
	ctx := context.Background()

	var stateValues []EnumValue
	state.ElementsAs(ctx, &stateValues, false)
	planned := make([]EnumValue, len(config))
	for i, c := range config {
		planned[i] = c
		if i < len(stateValues) {
			if c.ValueId.IsNull() {
				planned[i].ValueId = stateValues[i].ValueId
			}
			if c.Number.IsNull() {
				planned[i].Number = stateValues[i].Number
			}
		}
	}

	plan := enumList(t, planned...)
	req := planmodifier.ListRequest{ConfigValue: enumList(t, config...), PlanValue: plan, StateValue: state}
	resp := &planmodifier.ListResponse{PlanValue: plan}
	enumValuesPlanModifier{}.PlanModifyList(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}

	var result []EnumValue
	resp.PlanValue.ElementsAs(ctx, &result, false)
	return result
}

func expect(t *testing.T, v EnumValue, valueId, number types.String) {
	t.Helper()
	if !v.ValueId.Equal(valueId) || !v.Number.Equal(number) {
		t.Errorf("%s: expected value_id %s and number %s, got %s and %s", v.Value, valueId, number, v.ValueId, v.Number)
	}
}

func TestEnumValuesMatchByValueId(t *testing.T) {
	state := enumList(t,
		enumValue("Red", s("id-red"), s("red")),
		enumValue("Blue", s("id-blue"), s("blue")),
	)

	// Reordered, the number is taken from the value with the same ID
	result := modify(t, state,
		enumValue("Blue", s("id-blue"), null),
		enumValue("Red", s("id-red"), null),
	)
	expect(t, result[0], s("id-blue"), s("blue"))
	expect(t, result[1], s("id-red"), s("red"))
}

func TestEnumValuesMatchByNumber(t *testing.T) {
	state := enumList(t,
		enumValue("Red", s("id-red"), s("red")),
		enumValue("Blue", s("id-blue"), s("blue")),
	)

	// Reordered, a new value inserted and Red renamed
	result := modify(t, state,
		enumValue("Blue", null, s("blue")),
		enumValue("Green", null, s("green")),
		enumValue("Rood", null, s("red")),
	)
	expect(t, result[0], s("id-blue"), s("blue"))
	expect(t, result[1], types.StringUnknown(), s("green"))
	expect(t, result[2], s("id-red"), s("red"))
}

// Labels don't have to be unique, two values with the same label must keep
// their own ID.
func TestEnumValuesDuplicateLabels(t *testing.T) {
	state := enumList(t,
		enumValue(`20" Wheel`, s("id-1"), s("wheel-a")),
		enumValue(`20" Wheel`, s("id-2"), s("wheel-b")),
	)

	result := modify(t, state,
		enumValue(`20" Wheel`, null, s("wheel-b")),
		enumValue(`20" Wheel`, null, s("wheel-a")),
	)
	expect(t, result[0], s("id-2"), s("wheel-b"))
	expect(t, result[1], s("id-1"), s("wheel-a"))
}

func TestEnumValuesWithoutKeyAreNeverMatchedByLabel(t *testing.T) {
	state := enumList(t,
		enumValue("Red", s("id-red"), null),
		enumValue("Blue", s("id-blue"), null),
	)

	// Blue moved to the first position: without value_id or number it is not
	// matched by its label, so it is a new value.
	result := modify(t, state,
		enumValue("Blue", null, null),
		enumValue("Red", null, null),
	)
	expect(t, result[0], types.StringUnknown(), types.StringUnknown())
	expect(t, result[1], types.StringUnknown(), types.StringUnknown())
}

func TestEnumValuesUnchangedWithoutKey(t *testing.T) {
	state := enumList(t,
		enumValue("Red", s("id-red"), s("red")),
		enumValue("Blue", s("id-blue"), s("blue")),
	)

	// Unchanged values without a configured key stay as they are, so a plan
	// without changes is empty. A value appended at the end is new.
	result := modify(t, state,
		enumValue("Red", null, null),
		enumValue("Blue", null, null),
		enumValue("Green", null, null),
	)
	expect(t, result[0], s("id-red"), s("red"))
	expect(t, result[1], s("id-blue"), s("blue"))
	expect(t, result[2], types.StringUnknown(), types.StringUnknown())
}
