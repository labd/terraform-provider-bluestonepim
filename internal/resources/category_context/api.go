package category_context

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/category"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// GetCategoryContext reads the values stored for the context, without falling
// back to the values of other contexts. Returns nil when the category does not
// exist.
func GetCategoryContext(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, contextId string) (*CategoryContext, diag.Diagnostic) {
	resp, err := client.GetNodeWithResponse(ctx, categoryId, &pim.GetNodeParams{
		Context:         utils.Ref(contextId),
		ContextFallback: utils.Ref(false),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	return &CategoryContext{
		CategoryId:  types.StringValue(categoryId),
		ContextId:   types.StringValue(contextId),
		Name:        nullIfEmpty(resp.JSON200.Name),
		Description: nullIfEmpty(resp.JSON200.Description),
	}, nil
}

// SetCategoryContext writes the changed values for the context. A null value
// clears the value in the context, so it falls back to the fallback context.
func SetCategoryContext(ctx context.Context, client pim.ClientWithResponsesInterface, current *CategoryContext, planned *CategoryContext) (*CategoryContext, diag.Diagnostic) {
	body := pim.PartialUpdateCatalogNodeJSONRequestBody{}
	if !planned.Name.Equal(current.Name) {
		body.Name = &pim.PropertyUpdateString{Value: planned.Name.ValueStringPointer()}
	}
	if !planned.Description.Equal(current.Description) {
		body.Description = &pim.PropertyUpdateString{Value: planned.Description.ValueStringPointer()}
	}

	categoryId := planned.CategoryId.ValueString()
	contextId := planned.ContextId.ValueString()

	if body.Name != nil || body.Description != nil {
		if d := category.PartialUpdate(ctx, client, categoryId, utils.Ref(contextId), body); d != nil {
			return nil, d
		}
	}

	result, d := GetCategoryContext(ctx, client, categoryId, contextId)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Category not found", "Category "+categoryId+" was not found")
	}
	return result, nil
}

func nullIfEmpty(v *string) types.String {
	if v == nil || *v == "" {
		return types.StringNull()
	}
	return types.StringValue(*v)
}
