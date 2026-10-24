package category_matrix_attribute

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

type valuesMap = map[string]map[string]string

func fromResponse(ctx context.Context, r *pim.CategoryAttributeMetadataResponse, categoryId, attributeId string) (*CategoryMatrixAttribute, diag.Diagnostic) {
	flags := cla.FlagsFromResponse(r, categoryId)
	result := &CategoryMatrixAttribute{
		CategoryId:            types.StringValue(categoryId),
		AttributeDefinitionId: types.StringValue(attributeId),
		Values:                types.MapNull(valuesType),
		Mandatory:             flags.Mandatory,
		Propagate:             flags.Propagate,
	}

	// Empty values are stored as null, the schema doesn't allow empty values
	if r.Matrix != nil && len(*r.Matrix) > 0 {
		values, diags := types.MapValueFrom(ctx, valuesType, *r.Matrix)
		if diags.HasError() {
			return nil, diags.Errors()[0]
		}
		result.Values = values
	}
	return result, nil
}

func toValues(ctx context.Context, v types.Map) (*valuesMap, diag.Diagnostic) {
	// The API requires the values to be present, an empty map clears them
	result := valuesMap{}
	if !v.IsNull() && !v.IsUnknown() {
		if diags := v.ElementsAs(ctx, &result, false); diags.HasError() {
			return nil, diags.Errors()[0]
		}
	}
	return &result, nil
}

// GetCategoryMatrixAttribute returns nil when the attribute is not assigned to the category.
func GetCategoryMatrixAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*CategoryMatrixAttribute, diag.Diagnostic) {
	r, d := cla.Get(ctx, client, categoryId, attributeId)
	if d != nil || r == nil {
		return nil, d
	}
	return fromResponse(ctx, r, categoryId, attributeId)
}

func CreateCategoryMatrixAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, planned *CategoryMatrixAttribute) (*CategoryMatrixAttribute, diag.Diagnostic) {
	categoryId := planned.CategoryId.ValueString()
	attributeId := planned.AttributeDefinitionId.ValueString()

	// Assigning a compound attribute also assigns its sub attributes, so the
	// attribute can already be assigned. Take it over instead of failing.
	existing, d := GetCategoryMatrixAttribute(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	if existing != nil {
		return UpdateCategoryMatrixAttribute(ctx, client, existing, planned)
	}

	values, d := toValues(ctx, planned.Values)
	if d != nil {
		return nil, d
	}

	response, err := client.CreateCatalogNodeAttributeMatrixWithResponse(ctx,
		categoryId, attributeId,
		&pim.CreateCatalogNodeAttributeMatrixParams{ForceCla: utils.Ref(true)},
		pim.CreateCatalogNodeAttributeMatrixJSONRequestBody{Matrix: values},
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to assign attribute definition to category", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
		return nil, d
	}

	// The flags cannot be set on create
	current := *planned
	current.Mandatory = types.BoolValue(false)
	current.Propagate = types.BoolValue(false)
	return UpdateCategoryMatrixAttribute(ctx, client, &current, planned)
}

// UpdateCategoryMatrixAttribute updates the values and flags in a single request.
func UpdateCategoryMatrixAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, current, planned *CategoryMatrixAttribute) (*CategoryMatrixAttribute, diag.Diagnostic) {
	categoryId := current.CategoryId.ValueString()
	attributeId := current.AttributeDefinitionId.ValueString()

	body := pim.UpdateNodeAttributeMatrixValueJSONRequestBody{}
	changed := cla.ApplyChangedFlags(current.flags(), planned.flags(), &body.Mandatory, &body.CopyAttribute, nil)

	if changed || !planned.Values.Equal(current.Values) {
		// The endpoint expects the complete set of values
		values, d := toValues(ctx, planned.Values)
		if d != nil {
			return nil, d
		}
		body.Matrix = values

		response, err := client.UpdateNodeAttributeMatrixValueWithResponse(ctx, categoryId, attributeId,
			&pim.UpdateNodeAttributeMatrixValueParams{ForceCla: utils.Ref(true)}, body)
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update category attribute", err.Error())
		}

		if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
			return nil, d
		}
	}

	r, d := cla.MustGet(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	return fromResponse(ctx, r, categoryId, attributeId)
}
