package compound_attribute_definition

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// GetCompoundAttributeDefinitionByID returns nil when the definition does not exist.
func GetCompoundAttributeDefinitionByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*CompoundAttributeDefinition, diag.Diagnostic) {
	resp, err := client.GetCompoundDefinitionWithResponse(ctx, id, &pim.GetCompoundDefinitionParams{})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	return fromResponse(resp.JSON200), nil
}

func CreateCompoundAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, resource *CompoundAttributeDefinition) (*CompoundAttributeDefinition, diag.Diagnostic) {
	resp, err := client.CreateCompoundDefinitionWithResponse(ctx,
		&pim.CreateCompoundDefinitionParams{Validation: utils.Ref(pim.CreateCompoundDefinitionParamsValidationNAME)},
		resource.createDto(),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create compound attribute definition", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")

	return getExisting(ctx, client, resourceId)
}

func UpdateCompoundAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, current *CompoundAttributeDefinition, planned *CompoundAttributeDefinition) (*CompoundAttributeDefinition, diag.Diagnostic) {
	resp, err := client.UpdateCompoundDefinitionWithResponse(ctx, current.Id.ValueString(),
		&pim.UpdateCompoundDefinitionParams{Validation: utils.Ref(pim.UpdateCompoundDefinitionParamsValidationNAME)},
		planned.toDto(),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to update compound attribute definition", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	if !planned.Description.Equal(current.Description) {
		if d := attribute_definition.UpdateDescription(ctx, client, current.Id.ValueString(), planned.Description); d != nil {
			return nil, d
		}
	}

	return getExisting(ctx, client, current.Id.ValueString())
}

func DeleteCompoundAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, id string) diag.Diagnostic {
	resp, err := client.DeleteCompoundDefinitionWithResponse(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete compound attribute definition", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil
	}

	return utils.AssertSuccessStatusCode(resp)
}

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*CompoundAttributeDefinition, diag.Diagnostic) {
	result, d := GetCompoundAttributeDefinitionByID(ctx, client, id)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Attribute definition not found", "Attribute definition "+id+" was not found after writing it")
	}
	return result, nil
}
