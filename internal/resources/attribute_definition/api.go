package attribute_definition

import (
	"context"
	"fmt"
	"github.com/labd/bluestonepim-go-sdk/pim"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetAttributeDefinitionByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*AttributeDefinition, diag.Diagnostic) {
	response, err := client.FindFilteredAttributeDefinitionsWithResponse(ctx, nil, pim.FindFilteredAttributeDefinitionsJSONRequestBody{
		Filters: utils.Ref([]pim.AttributeDefinitionFilterDto{
			{
				Type:   utils.Ref[pim.AttributeDefinitionFilterDtoType]("ID_IN"),
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
		Id:          types.StringPointerValue(resource.Id),
		Name:        types.StringValue(resource.Name),
		Number:      types.StringPointerValue(resource.Number),
		DataType:    types.StringValue(string(*resource.DataType)),
		ContentType: types.StringPointerValue(resource.ContentType),
		Unit:        types.StringPointerValue(resource.Unit),
	}
	return result, nil
}

func UpdateAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, current *AttributeDefinition, resource *AttributeDefinition) (*AttributeDefinition, diag.Diagnostic) {
	if !(resource.Name.Equal(current.Name) && resource.Number.Equal(current.Number)) {
		res, err := client.UpdateMetadataWithResponse(ctx, current.Id.ValueString(),
			&pim.UpdateMetadataParams{},
			pim.UpdateMetadataJSONRequestBody{
				Name: &pim.PropertyUpdateString{
					Value: resource.Name.ValueStringPointer(),
				},
				Number: &pim.PropertyUpdateString{
					Value: resource.Number.ValueStringPointer(),
				},
			})

		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update attribute definition", err.Error())
		}

		if d := utils.AssertStatusCode(res, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	return GetAttributeDefinitionByID(ctx, client, current.Id.ValueString())
}

func CreateAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, resource *AttributeDefinition) (*AttributeDefinition, diag.Diagnostic) {
	response, err := client.CreateAttributeDefinitionWithResponse(ctx,
		&pim.CreateAttributeDefinitionParams{
			Validation: utils.Ref[pim.CreateAttributeDefinitionParamsValidation]("NAME"),
		},
		pim.CreateAttributeDefinitionJSONRequestBody{
			Charset:        resource.CharacterSet.ValueStringPointer(),
			ContentType:    resource.ContentType.ValueStringPointer(),
			DataType:       utils.Ref(pim.SimpleAttributeDefinitionRequestDataType(resource.DataType.ValueString())),
			ExternalSource: resource.ExternalSource.ValueBoolPointer(),
			GroupId:        resource.GroupID.ValueStringPointer(),
			Name:           resource.Name.ValueString(),
			Number:         resource.Number.ValueStringPointer(),
			//TODO: set restrictions
			Restrictions: nil,
			Unit:         resource.Unit.ValueStringPointer(),
		},
	)

	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create attribute definition", err.Error())
		return nil, d
	}

	if d := utils.AssertStatusCode(response, http.StatusCreated); d != nil {
		return nil, d
	}

	resourceId := response.HTTPResponse.Header.Get("Resource-Id")
	return GetAttributeDefinitionByID(ctx, client, resourceId)
}

func DeleteAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, id string) diag.Diagnostic {
	response, err := client.DeleteAttributeDefinitionWithResponse(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete attribute definition", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
		return d
	}

	return nil
}
