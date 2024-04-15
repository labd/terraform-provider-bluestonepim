package attribute

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetAttributeDefinitionByID(ctx context.Context, client *pim.ClientWithResponses, id string) (*AttributeDefinition, diag.Diagnostic) {
	response, err := client.FindFilteredWithResponse(ctx, nil, pim.FindFilteredJSONRequestBody{
		Filters: utils.Ref([]pim.AttributeDefinitionFilterDto{
			{
				Type:   utils.Ref(pim.AttributeDefinitionFilterDtoType("ID_IN")),
				Values: utils.Ref([]string{id}),
			},
		}),
	})

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
	if len(resources) != 1 {
		return nil, diag.NewErrorDiagnostic(
			"Unexpected data",
			fmt.Sprintf("Expected 1 catalog, got %d", len(resources)))
	}

	resource := resources[0]

	result := &AttributeDefinition{
		Id:          utils.NewStringValue(resource.Id),
		Name:        types.StringValue(resource.Name),
		Number:      utils.NewStringValue(resource.Number),
		DataType:    types.StringValue(string(*resource.DataType)),
		ContentType: utils.NewStringValue(resource.ContentType),
	}
	return result, nil
}

func UpdateAttributeDefinition(ctx context.Context, client *pim.ClientWithResponses, current *AttributeDefinition, resource *AttributeDefinition) (*AttributeDefinition, diag.Diagnostic) {

	return GetAttributeDefinitionByID(ctx, client, current.Id.ValueString())
}

func CreateAttributeDefinition(ctx context.Context, client *pim.ClientWithResponses, resource *AttributeDefinition) (*AttributeDefinition, diag.Diagnostic) {
	response, err := client.Create2WithResponse(ctx,
		&pim.Create2Params{
			Validation: utils.Ref(pim.Create2ParamsValidation("NAME")),
		},
		pim.Create2JSONRequestBody{
			Name:        resource.Name.ValueString(),
			Number:      utils.OptionalValueString(resource.Number),
			DataType:    utils.Ref(pim.SimpleAttributeDefinitionRequestDataType(resource.DataType.ValueString())),
			ContentType: resource.ContentType.ValueStringPointer(),
		},
	)

	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create attribute definition", err.Error())
		return nil, d
	}

	if d := utils.AssertStatusCode(response, http.StatusCreated); d != nil {
		return nil, d
	}

	if err != nil {
		d := diag.NewErrorDiagnostic("Error creating attribute definition", err.Error())
		return nil, d
	}

	resourceId := response.HTTPResponse.Header.Get("Resource-Id")
	return GetAttributeDefinitionByID(ctx, client, resourceId)
}

func DeleteAttributeDefinition(ctx context.Context, client *pim.ClientWithResponses, id string) diag.Diagnostic {
	response, err := client.Delete2WithResponse(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete attribute definition", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
		return d
	}

	return nil
}
