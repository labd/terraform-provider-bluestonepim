package attribute_definition

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// UpdateDescription sets the description of any attribute definition. After
// create the type specific update endpoints don't change the description, it
// has to go through the metadata endpoint.
func UpdateDescription(ctx context.Context, client pim.ClientWithResponsesInterface, id string, description types.String) diag.Diagnostic {
	resU, err := client.UpdateMetadataWithResponse(ctx, id, &pim.UpdateMetadataParams{},
		pim.UpdateMetadataJSONRequestBody{
			Description: &pim.PropertyUpdateString{
				Value: description.ValueStringPointer(),
			},
		})
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to update attribute definition description", err.Error())
	}
	return utils.AssertStatusCode(resU, http.StatusNoContent)
}

// IdentifiableValue is a column or row of a column / matrix attribute definition.
type IdentifiableValue struct {
	Id    types.String `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
}

func IdentifiableValuesSchema(description string, required bool) schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: description,
		Required:            required,
		Optional:            !required,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "Platform-generated identifier.",
					Computed:            true,
				},
				"value": schema.StringAttribute{
					MarkdownDescription: "The label.",
					Required:            true,
				},
			},
		},
	}
}

// ToIdentifiableValueDtos converts the planned values to DTOs. Values without a
// known ID (e.g. because the plan marked them as unknown) get the ID of the
// value with the same label in the current state, so existing columns / rows
// keep their identity.
func ToIdentifiableValueDtos(planned []IdentifiableValue, current []IdentifiableValue) *[]pim.IdentifiableValueDto {
	known := make(map[string]string, len(current))
	for _, v := range current {
		if !v.Id.IsNull() && !v.Id.IsUnknown() {
			known[v.Value.ValueString()] = v.Id.ValueString()
		}
	}

	result := make([]pim.IdentifiableValueDto, 0, len(planned))
	for _, v := range planned {
		var id *string
		if !v.Id.IsNull() && !v.Id.IsUnknown() {
			id = v.Id.ValueStringPointer()
		} else if existing, ok := known[v.Value.ValueString()]; ok {
			id = utils.Ref(existing)
		}

		result = append(result, pim.IdentifiableValueDto{
			Id:    id,
			Value: v.Value.ValueStringPointer(),
		})
	}
	return &result
}

func FromIdentifiableValueDtos(values *[]pim.IdentifiableValueDto) []IdentifiableValue {
	if values == nil {
		return nil
	}
	result := make([]IdentifiableValue, 0, len(*values))
	for _, v := range *values {
		result = append(result, IdentifiableValue{
			Id:    types.StringPointerValue(v.Id),
			Value: types.StringPointerValue(v.Value),
		})
	}
	return result
}

// CommonSchemaAttributes returns the schema attributes shared by the
// column, matrix and dictionary attribute definition resources.
func CommonSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			MarkdownDescription: "Platform-generated unique identifier of the attribute definition.",
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "The name of the attribute definition.",
			Required:            true,
		},
		"number": schema.StringAttribute{
			MarkdownDescription: "The number of the attribute definition.",
			Optional:            true,
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "The description of the attribute definition.",
			Optional:            true,
		},
		"external_source": schema.BoolAttribute{
			MarkdownDescription: "Whether the attribute is an external source.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"internal": schema.BoolAttribute{
			MarkdownDescription: "Whether the attribute is internal.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"group_id": schema.StringAttribute{
			MarkdownDescription: "The ID of the attribute group (see `bluestonepim_attribute_group`).",
			Optional:            true,
		},
		"context_aware": ContextAwareSchema(),
	}
}
