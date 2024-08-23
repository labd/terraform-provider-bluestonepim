package category

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetCategoryByID(ctx context.Context, client *pim.ClientWithResponses, id string) (*Category, diag.Diagnostic) {
	response, err := client.GetFilteredNodesWithResponse(ctx, nil, pim.GetFilteredNodesJSONRequestBody{
		Filters: utils.Ref([]pim.CategoryFilter{
			{
				Type:   utils.Ref(pim.CategoryFilterType("ID_IN")),
				Values: utils.Ref([]string{id}),
			},
		}),
	})

	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if response.StatusCode() != http.StatusOK {
		return nil, diag.NewErrorDiagnostic("Unexpected status code", fmt.Sprintf("Expected 200, got %d", response.StatusCode()))
	}

	categories := *response.JSON200.Data
	if (len(categories)) == 0 {
		return nil, nil
	}

	if len(categories) != 1 {
		return nil, diag.NewErrorDiagnostic(
			"Unexpected data",
			fmt.Sprintf("Expected 1 catalog, got %d", len(categories)))
	}

	category := categories[0]

	resource := &Category{
		Id:       utils.NewStringValue(category.Id),
		Name:     utils.NewStringValue(category.Name),
		Number:   utils.NewStringValue(category.Number),
		ParentId: utils.NewStringValue(category.ParentId),
	}
	return resource, nil
}

func UpdateCategory(ctx context.Context, client *pim.ClientWithResponses, current *Category, resource *Category) (*Category, diag.Diagnostic) {
	if (resource.Name.ValueString() != current.Name.ValueString()) ||
		(resource.Number.ValueString() != current.Number.ValueString()) {
		response, err := client.UpdateCatalogNodeWithResponse(ctx, resource.Id.ValueString(), nil,
			pim.UpdateCatalogNodeJSONRequestBody{
				Name:   resource.Name.ValueString(),
				Number: utils.OptionalValueString(resource.Number),
			})

		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update category", err.Error())
		}

		if d := utils.AssertStatusCode(response, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	if resource.ParentId.ValueString() != current.ParentId.ValueString() {
		response, err := client.MoveCatalogNodeWithResponse(ctx, resource.Id.ValueString(), pim.MoveCatalogNodeJSONRequestBody{
			ParentId: utils.OptionalValueString(resource.ParentId),
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

func CreateCategory(ctx context.Context, client *pim.ClientWithResponses, resource *Category) (*Category, diag.Diagnostic) {
	response, err := client.CreateCategoryWithResponse(ctx,
		&pim.CreateCategoryParams{
			Validation: "NAME",
		},
		pim.CreateCategoryJSONRequestBody{
			Name:     resource.Name.ValueString(),
			Number:   utils.OptionalValueString(resource.Number),
			ParentId: utils.OptionalValueString(resource.ParentId),
		},
	)

	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create category", err.Error())
		return nil, d
	}

	if d := utils.AssertStatusCode(response, http.StatusCreated); d != nil {
		return nil, d
	}

	resourceId := response.HTTPResponse.Header.Get("Resource-Id")
	return GetCategoryByID(ctx, client, resourceId)
}

func DeleteCategory(ctx context.Context, client *pim.ClientWithResponses, resource *Category) diag.Diagnostic {
	response, err := client.DeleteCategoryNodeWithResponse(ctx, resource.Id.ValueString())
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete category", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusNoContent); d != nil {
		return d
	}

	return nil
}
