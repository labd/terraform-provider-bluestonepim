package attribute_definition

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The restrictions each data type supports, see
// https://help.bluestonepim.com/attribute-types-1 and the API reference.
var (
	enumDataTypes  = []string{"single_select", "multi_select"}
	rangeDataTypes = []string{"integer", "decimal", "date", "time", "date_time"}
	textDataTypes  = []string{"text", "multiline", "formatted_text", "pattern"}
)

// ValidateConfig rejects restrictions that don't apply to the data type at
// plan time. The API rejects them at apply with "Invalid data type or
// restriction definition is invalid for the given data type".
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var dataType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("data_type"), &dataType)...)
	if resp.Diagnostics.HasError() || dataType.IsNull() || dataType.IsUnknown() {
		return
	}
	resp.Diagnostics.Append(validateRestrictions(ctx, req.Config, dataType.ValueString())...)
}

func validateRestrictions(ctx context.Context, config tfsdk.Config, dataType string) diag.Diagnostics {
	var diags diag.Diagnostics

	isSet := func(p path.Path) bool {
		var v attr.Value
		diags.Append(config.GetAttribute(ctx, p, &v)...)
		return v != nil && !v.IsNull()
	}
	requireDataType := func(p path.Path, allowed []string) {
		if isSet(p) && !slices.Contains(allowed, dataType) {
			diags.AddAttributeError(p, "Restriction not supported for data type",
				fmt.Sprintf("`%s` can only be used with the data types %s, not with `%s`.",
					attributeName(p), strings.Join(allowed, ", "), dataType))
		}
	}

	restrictions := path.Root("restrictions")
	requireDataType(path.Root("content_type"), []string{"formatted_text"})
	requireDataType(restrictions.AtName("enum"), enumDataTypes)
	requireDataType(restrictions.AtName("range"), rangeDataTypes)
	requireDataType(restrictions.AtName("text"), textDataTypes)
	// Regular expressions are a separate data type, the text restriction of
	// the other text types only supports max_length and whitespaces.
	requireDataType(restrictions.AtName("text").AtName("pattern"), []string{"pattern"})

	if slices.Contains(enumDataTypes, dataType) {
		var values types.List
		diags.Append(config.GetAttribute(ctx, restrictions.AtName("enum").AtName("values"), &values)...)
		if values.IsNull() || (!values.IsUnknown() && len(values.Elements()) == 0) {
			diags.AddAttributeError(restrictions.AtName("enum").AtName("values"), "Missing select values",
				fmt.Sprintf("A `%s` attribute definition needs at least one value in `restrictions.enum.values`.", dataType))
		}
	}

	return diags
}

func attributeName(p path.Path) string {
	return strings.ReplaceAll(strings.TrimPrefix(p.String(), "restrictions."), "]", "")
}

// dataTypeImmutable rejects changing the data type of an existing definition.
// Replacing the definition instead would remove the attribute and its values
// from every product.
type dataTypeImmutable struct{}

func (m dataTypeImmutable) Description(_ context.Context) string {
	return "The data type of an existing attribute definition cannot be changed."
}

func (m dataTypeImmutable) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m dataTypeImmutable) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// A create (also part of a deliberate `terraform apply -replace`) has no state
	if req.State.Raw.IsNull() || req.StateValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if !req.PlanValue.Equal(req.StateValue) {
		resp.Diagnostics.AddAttributeError(req.Path, "Data type cannot be changed",
			fmt.Sprintf("The data type of an existing attribute definition cannot be changed from `%s` to `%s`. "+
				"Replacing the definition would remove the attribute and its values from every product. Create a "+
				"new attribute definition and migrate the values instead. To replace it anyway, use "+
				"`terraform apply -replace=<address>`.",
				req.StateValue.ValueString(), req.PlanValue.ValueString()))
	}
}
