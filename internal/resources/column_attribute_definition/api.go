package column_attribute_definition

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

// GetColumnAttributeDefinitionByID returns nil when the definition does not exist.
func GetColumnAttributeDefinitionByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*ColumnAttributeDefinition, diag.Diagnostic) {
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
	if resource.DataType == nil || *resource.DataType != pim.AttributeDefinitionResponseDataTypeColumn {
		return nil, diag.NewErrorDiagnostic("Unexpected attribute definition type",
			fmt.Sprintf("Attribute definition %s is not a column definition", id))
	}

	result := &ColumnAttributeDefinition{
		Id:             types.StringPointerValue(resource.Id),
		Name:           types.StringPointerValue(resource.Name),
		Number:         types.StringPointerValue(resource.Number),
		Description:    types.StringPointerValue(resource.Description),
		ExternalSource: types.BoolPointerValue(resource.ExternalSource),
		GroupID:        types.StringPointerValue(resource.GroupId),
		Internal:       types.BoolPointerValue(resource.Internal),
		ContextAware:   types.BoolPointerValue(resource.ContextAware),
	}
	if resource.Restrictions != nil && resource.Restrictions.Column != nil {
		result.Columns = attribute_definition.FromIdentifiableValueDtos(resource.Restrictions.Column.Columns)
	}
	return result, nil
}

func toDto(planned *ColumnAttributeDefinition, current *ColumnAttributeDefinition) pim.ColumnAttributeDefinitionDto {
	var currentColumns []attribute_definition.IdentifiableValue
	if current != nil {
		currentColumns = current.Columns
	}

	return pim.ColumnAttributeDefinitionDto{
		ExternalSource: planned.ExternalSource.ValueBoolPointer(),
		GroupId:        planned.GroupID.ValueStringPointer(),
		Internal:       planned.Internal.ValueBoolPointer(),
		ContextAware:   utils.KnownBoolPointer(planned.ContextAware),
		Name:           planned.Name.ValueStringPointer(),
		Number:         utils.KnownStringPointer(planned.Number),
		Restrictions: &pim.ColumnRestrictionDto{
			Column: &pim.ColumnRestrictionsDto{
				Columns: attribute_definition.ToIdentifiableValueDtos(planned.Columns, currentColumns),
			},
		},
	}
}

func CreateColumnAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, resource *ColumnAttributeDefinition) (*ColumnAttributeDefinition, diag.Diagnostic) {
	resp, err := client.CreateColumnDefinitionWithResponse(ctx,
		&pim.CreateColumnDefinitionParams{Validation: utils.Ref(pim.CreateColumnDefinitionParamsValidationNAME)},
		toDto(resource, nil),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create column attribute definition", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")

	if !resource.Description.IsNull() {
		if d := attribute_definition.UpdateDescription(ctx, client, resourceId, resource.Description); d != nil {
			return nil, d
		}
	}

	return getExisting(ctx, client, resourceId)
}

func UpdateColumnAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, current *ColumnAttributeDefinition, planned *ColumnAttributeDefinition) (*ColumnAttributeDefinition, diag.Diagnostic) {
	resp, err := client.UpdateColumnDefinitionWithResponse(ctx, current.Id.ValueString(),
		&pim.UpdateColumnDefinitionParams{Validation: utils.Ref(pim.UpdateColumnDefinitionParamsValidationNAME)},
		toDto(planned, current),
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to update column attribute definition", err.Error())
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

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*ColumnAttributeDefinition, diag.Diagnostic) {
	result, d := GetColumnAttributeDefinitionByID(ctx, client, id)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Attribute definition not found", "Attribute definition "+id+" was not found after writing it")
	}
	return result, nil
}
