package category_attribute

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetCategoryAttributeByID(
	ctx context.Context,
	client *pim.ClientWithResponses,
	categoryId, attributeId string,
) (*CategoryAttribute, diag.Diagnostic) {
	response, err := client.ListAttributesAttachedToGivenNodeWithResponse(
		ctx, categoryId, &pim.ListAttributesAttachedToGivenNodeParams{})

	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusOK); d != nil {
		return nil, d
	}

	resources := *response.JSON200.Data
	if len(resources) == 0 {
		return nil, nil
	}

	for _, resource := range resources {
		if *resource.AttributeDefinitionId != attributeId {
			continue
		}

		// I think this should not be able to happen (since we fetch on category),
		// but we check it anyway.
		if *resource.AssignedOn != categoryId {
			continue
		}

		result := &CategoryAttribute{
			CategoryId:  types.StringValue(categoryId),
			AttributeId: types.StringValue(attributeId),
		}

		if resource.MandatorySetOn != nil {
			result.Mandatory = types.BoolValue(*resource.MandatorySetOn == categoryId)
		} else {
			result.Mandatory = types.BoolValue(false)
		}

		return result, nil
	}

	return nil, nil
}

func UpdateAttributeDefinition(
	ctx context.Context,
	client *pim.ClientWithResponses,
	current *CategoryAttribute,
	resource *CategoryAttribute,
) (*CategoryAttribute, diag.Diagnostic) {

	if resource.Mandatory.ValueBool() != current.Mandatory.ValueBool() {
		response, err := client.UpdateNodeAttributeValueWithResponse(ctx,
			current.CategoryId.ValueString(),
			current.AttributeId.ValueString(),
			nil,
			pim.UpdateNodeAttributeValueJSONRequestBody{
				Mandatory: resource.Mandatory.ValueBoolPointer(),
			},
		)
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update category", err.Error())
		}

		if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
			return nil, d
		}
	}

	return GetCategoryAttributeByID(
		ctx, client, current.CategoryId.ValueString(), current.AttributeId.ValueString())
}

func AssignAttributeDefinition(
	ctx context.Context,
	client *pim.ClientWithResponses,
	resource *CategoryAttribute,
) (*CategoryAttribute, diag.Diagnostic) {

	response, err := client.CreateCatalogNodeAttributeWithResponse(ctx,
		resource.CategoryId.ValueString(),
		resource.AttributeId.ValueString(),
		&pim.CreateCatalogNodeAttributeParams{
			ForceCla: utils.Ref(true),
		},
		pim.CreateCatalogNodeAttributeJSONRequestBody{},
	)

	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create attribute definition", err.Error())
		return nil, d
	}

	if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
		return nil, d
	}

	current := &CategoryAttribute{
		CategoryId:  resource.CategoryId,
		AttributeId: resource.AttributeId,
	}

	// We call the update here since the API doesn't support setting certain flags
	// on creation.
	return UpdateAttributeDefinition(ctx, client, current, resource)
}

func UnassignAttributeDefinition(
	ctx context.Context,
	client *pim.ClientWithResponses,
	categoryId, attributeId string,
) diag.Diagnostic {
	response, err := client.DeleteAttributeFromNodeWithResponse(ctx, categoryId, attributeId)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to remove attribute definition from category", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusNoContent); d != nil {
		return d
	}

	return nil
}
