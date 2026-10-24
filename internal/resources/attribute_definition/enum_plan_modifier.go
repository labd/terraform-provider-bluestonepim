package attribute_definition

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// enumValuesPlanModifier keeps the platform-generated `value_id` (and `number`)
// of select values stable. Products store the ID of the selected value, so a
// value must never get the ID of another value.
//
// Terraform matches nested list elements by index, so adding, removing or
// reordering values would shift the IDs onto other values. Instead, a value is
// matched to an existing value by:
//
//  1. its configured `value_id`, or
//  2. its configured `number`, which is unique within the definition.
//
// A value with neither is only kept as is when it is unchanged (same position
// and same label as in the state), otherwise it is a new value. There is no
// lookup by label, labels don't have to be unique.
type enumValuesPlanModifier struct{}

func (m enumValuesPlanModifier) Description(_ context.Context) string {
	return "Keeps the value_id and number of existing select values, matched by value_id or number."
}

func (m enumValuesPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m enumValuesPlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	var planned, config, state []EnumValue
	resp.Diagnostics.Append(req.PlanValue.ElementsAs(ctx, &planned, false)...)
	resp.Diagnostics.Append(req.ConfigValue.ElementsAs(ctx, &config, false)...)
	if !req.StateValue.IsNull() && !req.StateValue.IsUnknown() {
		resp.Diagnostics.Append(req.StateValue.ElementsAs(ctx, &state, false)...)
	}
	if resp.Diagnostics.HasError() || len(planned) != len(config) {
		return
	}

	byId := map[string]EnumValue{}
	byNumber := map[string]EnumValue{}
	for _, v := range state {
		if isKnown(v.ValueId) {
			byId[v.ValueId.ValueString()] = v
		}
		if isKnown(v.Number) {
			byNumber[v.Number.ValueString()] = v
		}
	}

	for i := range planned {
		var match EnumValue
		var found bool

		switch {
		case isKnown(config[i].ValueId):
			match, found = byId[config[i].ValueId.ValueString()]
		case isKnown(config[i].Number):
			match, found = byNumber[config[i].Number.ValueString()]
		case i < len(state) && state[i].Value.Equal(planned[i].Value):
			// Unchanged value without a configured key
			match, found = state[i], true
		}

		if config[i].ValueId.IsNull() {
			planned[i].ValueId = types.StringUnknown()
			if found {
				planned[i].ValueId = match.ValueId
			}
		}

		if config[i].Number.IsNull() {
			planned[i].Number = types.StringUnknown()
			if found {
				planned[i].Number = match.Number
			}
		}
	}

	result, diags := types.ListValueFrom(ctx, req.PlanValue.ElementType(ctx), planned)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.PlanValue = result
}

func isKnown(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown()
}
