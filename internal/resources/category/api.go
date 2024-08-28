package category

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetCategoryByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*Category, diag.Diagnostic) {
	resp, err := client.GetNodeWithResponse(ctx, id, nil)

	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
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
	return resource, nil
}

func UpdateCategory(ctx context.Context, client pim.ClientWithResponsesInterface, current *Category, planned *Category) (*Category, diag.Diagnostic) {
	if !(planned.Name.Equal(current.Name) && planned.Number.Equal(current.Number) && planned.Description.Equal(current.Description)) {
		response, err := client.UpdateCatalogNodeWithResponse(ctx, planned.Id.ValueString(), nil,
			pim.UpdateCatalogNodeJSONRequestBody{
				Name:        planned.Name.ValueString(),
				Number:      planned.Number.ValueStringPointer(),
				Description: planned.Description.ValueStringPointer(),
			})

		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update category", err.Error())
		}

		if d := utils.AssertStatusCode(response, http.StatusNoContent); d != nil {
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

	return GetCategoryByID(ctx, client, current.Id.ValueString())
}

func CreateCategory(ctx context.Context, client pim.ClientWithResponsesInterface, resource *Category) (*Category, diag.Diagnostic) {
	res, err := client.CreateCategoryWithResponse(ctx,
		&pim.CreateCategoryParams{
			Validation: "NAME",
		},
		pim.CreateCategoryJSONRequestBody{
			Name:     resource.Name.ValueString(),
			Number:   resource.Number.ValueStringPointer(),
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

	//Workaround because we cannot set description on create
	resU, err := client.UpdateCatalogNodeWithResponse(ctx, resourceId, nil, pim.UpdateCatalogNodeJSONRequestBody{
		Name:        resource.Name.ValueString(),
		Number:      resource.Number.ValueStringPointer(),
		Description: resource.Description.ValueStringPointer(),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed updating category", err.Error())
	}

	if d := utils.AssertStatusCode(resU, http.StatusNoContent); d != nil {
		return nil, d
	}

	return GetCategoryByID(ctx, client, resourceId)
}

func DeleteCategory(ctx context.Context, client pim.ClientWithResponsesInterface, resource *Category) diag.Diagnostic {
	response, err := client.DeleteCategoryNodeWithResponse(ctx, resource.Id.ValueString())
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete category", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusNoContent); d != nil {
		return d
	}

	return nil
}
