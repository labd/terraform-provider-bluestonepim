package category

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetCategoryByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*Category, diag.Diagnostic) {
	resp, err := client.GetNodeWithResponse(ctx, id, &pim.GetNodeParams{})

	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	resource := &Category{
		Id:          types.StringPointerValue(resp.JSON200.Id),
		Name:        types.StringPointerValue(resp.JSON200.Name),
		Number:      types.StringPointerValue(resp.JSON200.Number),
		ParentId:    types.StringPointerValue(resp.JSON200.ParentId),
		Description: types.StringPointerValue(resp.JSON200.Description),
	}

	// The API returns an empty description for categories without one
	if resource.Description.ValueString() == "" {
		resource.Description = types.StringNull()
	}
	return resource, nil
}

func UpdateCategory(ctx context.Context, client pim.ClientWithResponsesInterface, current *Category, planned *Category) (*Category, diag.Diagnostic) {
	// Partial update, so only the changed fields are sent
	body := pim.PartialUpdateCatalogNodeJSONRequestBody{}
	if !planned.Name.Equal(current.Name) {
		body.Name = &pim.PropertyUpdateString{Value: planned.Name.ValueStringPointer()}
	}
	if !planned.Number.IsUnknown() && !planned.Number.Equal(current.Number) {
		body.Number = planned.Number.ValueStringPointer()
	}
	if !planned.Description.Equal(current.Description) {
		body.Description = &pim.PropertyUpdateString{Value: planned.Description.ValueStringPointer()}
	}

	if body.Name != nil || body.Number != nil || body.Description != nil {
		if d := PartialUpdate(ctx, client, current.Id.ValueString(), nil, body); d != nil {
			return nil, d
		}
	}

	if !planned.ParentId.Equal(current.ParentId) {
		response, err := client.MoveCatalogNodeWithResponse(ctx, planned.Id.ValueString(), pim.MoveCatalogNodeJSONRequestBody{
			ParentId: planned.ParentId.ValueStringPointer(),
		})

		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update category", err.Error())
		}

		if d := utils.AssertStatusCode(response, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	result, d := GetCategoryByID(ctx, client, current.Id.ValueString())
	if d == nil && result == nil {
		d = utils.NotFoundError("Category", current.Id.ValueString())
	}
	return result, d
}

func CreateCategory(ctx context.Context, client pim.ClientWithResponsesInterface, resource *Category) (*Category, diag.Diagnostic) {
	res, err := client.CreateCategoryWithResponse(ctx,
		&pim.CreateCategoryParams{
			Validation: utils.Ref(pim.CreateCategoryParamsValidationNAME),
		},
		pim.CreateCategoryJSONRequestBody{
			Name:     resource.Name.ValueString(),
			Number:   utils.KnownStringPointer(resource.Number),
			ParentId: resource.ParentId.ValueStringPointer(),
		},
	)

	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create category", err.Error())
		return nil, d
	}

	if d := utils.AssertStatusCode(res, http.StatusCreated); d != nil {
		return nil, d
	}

	resourceId := res.HTTPResponse.Header.Get("Resource-Id")

	// The description cannot be set on create
	if !resource.Description.IsNull() {
		body := pim.PartialUpdateCatalogNodeJSONRequestBody{
			Description: &pim.PropertyUpdateString{Value: resource.Description.ValueStringPointer()},
		}
		if d := PartialUpdate(ctx, client, resourceId, nil, body); d != nil {
			return nil, d
		}
	}

	result, d := GetCategoryByID(ctx, client, resourceId)
	if d == nil && result == nil {
		d = utils.NotFoundError("Category", resourceId)
	}
	return result, d
}

func DeleteCategory(ctx context.Context, client pim.ClientWithResponsesInterface, resource *Category) diag.Diagnostic {
	response, err := client.DeleteCategoryNodeWithResponse(ctx, resource.Id.ValueString())
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete category", err.Error())
	}

	if utils.IsNotFound(response) {
		return nil
	}

	// The category is removed asynchronously
	return utils.AssertStatusCode(response, http.StatusAccepted)
}

// PartialUpdate updates only the fields set in the body. When contextId is set
// the values are written for that context only.
func PartialUpdate(ctx context.Context, client pim.ClientWithResponsesInterface, id string, contextId *string, body pim.PartialUpdateCatalogNodeJSONRequestBody) diag.Diagnostic {
	response, err := client.PartialUpdateCatalogNodeWithResponse(ctx, id,
		&pim.PartialUpdateCatalogNodeParams{Context: contextId}, body)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to update category", err.Error())
	}

	return utils.AssertStatusCode(response, http.StatusNoContent)
}
