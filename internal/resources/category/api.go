package category

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/labd/bluestone-pim-go-sdk/pim"

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

func CreateCategory(ctx context.Context, client *pim.ClientWithResponses, resource *Category) (*string, diag.Diagnostic) {
	response, err := client.Create8WithResponse(ctx,
		&pim.Create8Params{
			Validation: "NAME",
		},
		pim.Create8JSONRequestBody{
			Name:     resource.Name.ValueString(),
			Number:   utils.OptionalValueString(resource.Number),
			ParentId: utils.OptionalValueString(resource.ParentId),
		},
	)

	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create category", err.Error())
		return nil, d
	}

	if response.StatusCode() != http.StatusCreated {
		d := diag.NewErrorDiagnostic("Unexpected status code", fmt.Sprintf("Expected 201, got %d", response.StatusCode()))
		return nil, d
	}

	if err != nil {
		d := diag.NewErrorDiagnostic("Error creating subscription", err.Error())
		return nil, d
	}

	resourceId := response.HTTPResponse.Header.Get("Resource-Id")
	return &resourceId, nil
}
