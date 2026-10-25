package matrix_attribute_definition

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_definition"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// GetMatrixAttributeDefinitionByID returns nil when the definition does not exist.
func GetMatrixAttributeDefinitionByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*MatrixAttributeDefinition, diag.Diagnostic) {
	resp, err := client.GetAttributeDefinitionWithResponse(ctx, id, &pim.GetAttributeDefinitionParams{})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	resource := resp.JSON200
	if resource.DataType == nil || *resource.DataType != pim.AttributeDefinitionResponseDataTypeMatrix {
		return nil, diag.NewErrorDiagnostic("Unexpected attribute definition type",
			fmt.Sprintf("Attribute definition %s is not a matrix definition", id))
	}

	result := &MatrixAttributeDefinition{
		Id:             types.StringPointerValue(resource.Id),
		Name:           types.StringPointerValue(resource.Name),
		Number:         types.StringPointerValue(resource.Number),
		Description:    types.StringPointerValue(resource.Description),
		ExternalSource: types.BoolPointerValue(resource.ExternalSource),
		GroupID:        types.StringPointerValue(resource.GroupId),
		Internal:       types.BoolPointerValue(resource.Internal),
		ContextAware:   types.BoolPointerValue(resource.ContextAware),
	}
	if resource.Restrictions != nil && resource.Restrictions.Matrix != nil {
		result.Columns = attribute_definition.FromIdentifiableValueDtos(resource.Restrictions.Matrix.Columns)
		result.Rows = attribute_definition.FromIdentifiableValueDtos(resource.Restrictions.Matrix.Rows)
	}
	return result, nil
}

// createDto is the update request plus the description, which can only be set
// on create (afterwards it is updated with the metadata endpoint).
func createDto(resource *MatrixAttributeDefinition) pim.MatrixAttributeDefinitionDto {
	dto := toDto(resource, nil)
	dto.Description = resource.Description.ValueStringPointer()
	return dto
}

func toDto(planned *MatrixAttributeDefinition, current *MatrixAttributeDefinition) pim.MatrixAttributeDefinitionDto {
	var currentColumns, currentRows []attribute_definition.IdentifiableValue
	if current != nil {
		currentColumns = current.Columns
		currentRows = current.Rows
	}

	return pim.MatrixAttributeDefinitionDto{
		ExternalSource: planned.ExternalSource.ValueBoolPointer(),
		GroupId:        planned.GroupID.ValueStringPointer(),
		Internal:       planned.Internal.ValueBoolPointer(),
		ContextAware:   utils.KnownBoolPointer(planned.ContextAware),
		Name:           planned.Name.ValueStringPointer(),
		Number:         utils.KnownStringPointer(planned.Number),
		Restrictions: &pim.MatrixRestrictionDto{
			Matrix: &pim.MatrixRestrictionsDto{
				Columns: attribute_definition.ToIdentifiableValueDtos(planned.Columns, currentColumns),
				Rows:    attribute_definition.ToIdentifiableValueDtos(planned.Rows, currentRows),
			},
		},
	}
}

func CreateMatrixAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, resource *MatrixAttributeDefinition) (*MatrixAttributeDefinition, diag.Diagnostic) {
	resp, err := client.CreateMatrixDefinitionWithResponse(ctx,
		&pim.CreateMatrixDefinitionParams{Validation: utils.Ref(pim.CreateMatrixDefinitionParamsValidationNAME)},
		createDto(resource),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create matrix attribute definition", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")

	return getExisting(ctx, client, resourceId)
}

func UpdateMatrixAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, current *MatrixAttributeDefinition, planned *MatrixAttributeDefinition) (*MatrixAttributeDefinition, diag.Diagnostic) {
	resp, err := client.UpdateMatrixDefinitionWithResponse(ctx, current.Id.ValueString(),
		&pim.UpdateMatrixDefinitionParams{Validation: utils.Ref(pim.UpdateMatrixDefinitionParamsValidationNAME)},
		toDto(planned, current),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to update matrix attribute definition", err.Error())
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

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*MatrixAttributeDefinition, diag.Diagnostic) {
	result, d := GetMatrixAttributeDefinitionByID(ctx, client, id)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Attribute definition not found", "Attribute definition "+id+" was not found after writing it")
	}
	return result, nil
}
