package category_dictionary_attribute

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func fromResponse(ctx context.Context, r *pim.CategoryAttributeMetadataResponse, categoryId, attributeId string) (*CategoryDictionaryAttribute, diag.Diagnostic) {
	flags := cla.FlagsFromResponse(r, categoryId)
	result := &CategoryDictionaryAttribute{
		CategoryId:            types.StringValue(categoryId),
		AttributeDefinitionId: types.StringValue(attributeId),
		ValueIds:              types.SetNull(types.StringType),
		Mandatory:             flags.Mandatory,
		Propagate:             flags.Propagate,
		Locked:                flags.Locked,
	}

	// No selected values is stored as null, the schema doesn't allow an empty set
	if r.Dictionary != nil && len(*r.Dictionary) > 0 {
		valueIds, diags := types.SetValueFrom(ctx, types.StringType, *r.Dictionary)
		if diags.HasError() {
			return nil, diags.Errors()[0]
		}
		result.ValueIds = valueIds
	}
	return result, nil
}

// GetCategoryDictionaryAttribute returns nil when the attribute is not assigned to the category.
func GetCategoryDictionaryAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*CategoryDictionaryAttribute, diag.Diagnostic) {
	r, d := cla.Get(ctx, client, categoryId, attributeId)
	if d != nil || r == nil {
		return nil, d
	}
	return fromResponse(ctx, r, categoryId, attributeId)
}

func CreateCategoryDictionaryAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, planned *CategoryDictionaryAttribute) (*CategoryDictionaryAttribute, diag.Diagnostic) {
	categoryId := planned.CategoryId.ValueString()
	attributeId := planned.AttributeDefinitionId.ValueString()

	// Assigning a compound attribute also assigns its sub attributes, so the
	// attribute can already be assigned. Take it over instead of failing.
	existing, d := GetCategoryDictionaryAttribute(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	if existing != nil {
		return UpdateCategoryDictionaryAttribute(ctx, client, existing, planned)
	}

	// Setting the values also assigns the attribute, so the explicit assign is
	// only needed without values.
	if planned.ValueIds.IsNull() {
		response, err := client.CreateCatalogNodeAttributeDictionaryWithResponse(ctx, categoryId, attributeId, &pim.CreateCatalogNodeAttributeDictionaryParams{})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to assign attribute definition to category", err.Error())
		}

		if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
			return nil, d
		}
	} else if d := setValues(ctx, client, categoryId, attributeId, planned.ValueIds); d != nil {
		return nil, d
	}

	// The flags cannot be set on create
	if d := cla.UpdateFlags(ctx, client, categoryId, attributeId, cla.Flags{}, planned.flags(), true); d != nil {
		return nil, d
	}

	return mustGet(ctx, client, categoryId, attributeId)
}

func UpdateCategoryDictionaryAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, current, planned *CategoryDictionaryAttribute) (*CategoryDictionaryAttribute, diag.Diagnostic) {
	categoryId := current.CategoryId.ValueString()
	attributeId := current.AttributeDefinitionId.ValueString()

	// Values and flags are separate endpoints
	if !planned.ValueIds.Equal(current.ValueIds) {
		if d := setValues(ctx, client, categoryId, attributeId, planned.ValueIds); d != nil {
			return nil, d
		}
	}

	if d := cla.UpdateFlags(ctx, client, categoryId, attributeId, current.flags(), planned.flags(), true); d != nil {
		return nil, d
	}

	return mustGet(ctx, client, categoryId, attributeId)
}

// setValues replaces the selected values, an empty list clears the selection.
func setValues(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string, v types.Set) diag.Diagnostic {
	valueIds := []string{}
	if !v.IsNull() && !v.IsUnknown() {
		if diags := v.ElementsAs(ctx, &valueIds, false); diags.HasError() {
			return diags.Errors()[0]
		}
	}

	response, err := client.SetDictionaryAttributeValuesForCategoryWithResponse(ctx, categoryId, attributeId,
		&pim.SetDictionaryAttributeValuesForCategoryParams{ForceCla: utils.Ref(true)},
		pim.SetDictionaryAttributeValuesForCategoryJSONRequestBody{ValueIds: &valueIds},
	)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to set dictionary values", err.Error())
	}

	return utils.AssertStatusCode(response, http.StatusAccepted)
}

func mustGet(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*CategoryDictionaryAttribute, diag.Diagnostic) {
	r, d := cla.MustGet(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	return fromResponse(ctx, r, categoryId, attributeId)
}
