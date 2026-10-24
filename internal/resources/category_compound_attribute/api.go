package category_compound_attribute

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func fromResponse(r *pim.CategoryAttributeMetadataResponse, categoryId, attributeId string) *CategoryCompoundAttribute {
	flags := cla.FlagsFromResponse(r, categoryId)
	return &CategoryCompoundAttribute{
		CategoryId:            types.StringValue(categoryId),
		AttributeDefinitionId: types.StringValue(attributeId),
		Mandatory:             flags.Mandatory,
		Propagate:             flags.Propagate,
	}
}

// GetCategoryCompoundAttribute returns nil when the attribute is not assigned to the category.
func GetCategoryCompoundAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*CategoryCompoundAttribute, diag.Diagnostic) {
	r, d := cla.Get(ctx, client, categoryId, attributeId)
	if d != nil || r == nil {
		return nil, d
	}
	return fromResponse(r, categoryId, attributeId), nil
}

// CreateCategoryCompoundAttribute assigns the compound attribute, which also
// assigns its sub attributes. Their values are managed with the resources
// matching their type.
func CreateCategoryCompoundAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, planned *CategoryCompoundAttribute) (*CategoryCompoundAttribute, diag.Diagnostic) {
	categoryId := planned.CategoryId.ValueString()
	attributeId := planned.AttributeDefinitionId.ValueString()

	response, err := client.CreateCatalogNodeAttributeCompoundWithResponse(ctx, categoryId, attributeId)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to assign attribute definition to category", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
		return nil, d
	}

	// The flags cannot be set on create
	return UpdateCategoryCompoundAttribute(ctx, client, &CategoryCompoundAttribute{
		CategoryId:            planned.CategoryId,
		AttributeDefinitionId: planned.AttributeDefinitionId,
	}, planned)
}

func UpdateCategoryCompoundAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, current, planned *CategoryCompoundAttribute) (*CategoryCompoundAttribute, diag.Diagnostic) {
	categoryId := current.CategoryId.ValueString()
	attributeId := current.AttributeDefinitionId.ValueString()

	if d := cla.UpdateFlags(ctx, client, categoryId, attributeId, current.flags(), planned.flags(), false); d != nil {
		return nil, d
	}

	r, d := cla.MustGet(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	return fromResponse(r, categoryId, attributeId), nil
}
