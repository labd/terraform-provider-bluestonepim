package dictionary_attribute_definition

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

// GetDictionaryAttributeDefinitionByID returns nil when the definition does not exist.
func GetDictionaryAttributeDefinitionByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*DictionaryAttributeDefinition, diag.Diagnostic) {
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
	if resource.DataType == nil || *resource.DataType != pim.AttributeDefinitionResponseDataTypeDictionary {
		return nil, diag.NewErrorDiagnostic("Unexpected attribute definition type",
			fmt.Sprintf("Attribute definition %s is not a dictionary definition", id))
	}

	return &DictionaryAttributeDefinition{
		Id:                 types.StringPointerValue(resource.Id),
		Name:               types.StringPointerValue(resource.Name),
		Number:             types.StringPointerValue(resource.Number),
		Description:        types.StringPointerValue(resource.Description),
		ExternalSource:     types.BoolPointerValue(resource.ExternalSource),
		GroupID:            types.StringPointerValue(resource.GroupId),
		Internal:           types.BoolPointerValue(resource.Internal),
		SelectedValueLimit: types.Int32PointerValue(resource.SelectedValuesLimit),
		ParentDefinitionId: types.StringPointerValue(resource.FilterParentDefinitionId),
		ContextAware:       types.BoolPointerValue(resource.ContextAware),
	}, nil
}

func toDto(planned *DictionaryAttributeDefinition) pim.DictionaryAttributeDefinitionDto {
	return pim.DictionaryAttributeDefinitionDto{
		ExternalSource:      planned.ExternalSource.ValueBoolPointer(),
		GroupId:             planned.GroupID.ValueStringPointer(),
		Internal:            planned.Internal.ValueBoolPointer(),
		Name:                planned.Name.ValueStringPointer(),
		Number:              utils.KnownStringPointer(planned.Number),
		SelectedValuesLimit: planned.SelectedValueLimit.ValueInt32Pointer(),
	}
}

func CreateDictionaryAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, resource *DictionaryAttributeDefinition) (*DictionaryAttributeDefinition, diag.Diagnostic) {
	resp, err := client.CreateDictionaryDefinitionWithResponse(ctx,
		&pim.CreateDictionaryDefinitionParams{Validation: utils.Ref(pim.CreateDictionaryDefinitionParamsValidationNAME)},
		pim.CreateDictionaryDefinitionJSONRequestBody{
			Description:         resource.Description.ValueStringPointer(),
			ExternalSource:      resource.ExternalSource.ValueBoolPointer(),
			GroupId:             resource.GroupID.ValueStringPointer(),
			Internal:            resource.Internal.ValueBoolPointer(),
			Name:                resource.Name.ValueStringPointer(),
			Number:              utils.KnownStringPointer(resource.Number),
			SelectedValuesLimit: resource.SelectedValueLimit.ValueInt32Pointer(),
			ContextAware:        utils.KnownBoolPointer(resource.ContextAware),
		},
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create dictionary attribute definition", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")

	// The parent cannot be set on create
	current := &DictionaryAttributeDefinition{
		Id:                 types.StringValue(resourceId),
		Description:        resource.Description,
		SelectedValueLimit: resource.SelectedValueLimit,
		ParentDefinitionId: types.StringNull(),
	}
	if d := updateExtras(ctx, client, current, resource); d != nil {
		return nil, d
	}

	return getExisting(ctx, client, resourceId)
}

func UpdateDictionaryAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, current *DictionaryAttributeDefinition, planned *DictionaryAttributeDefinition) (*DictionaryAttributeDefinition, diag.Diagnostic) {
	hasChanges := !(planned.Name.Equal(current.Name) &&
		planned.Number.Equal(current.Number) &&
		planned.ExternalSource.Equal(current.ExternalSource) &&
		planned.Internal.Equal(current.Internal) &&
		planned.GroupID.Equal(current.GroupID) &&
		planned.SelectedValueLimit.Equal(current.SelectedValueLimit))

	if hasChanges {
		resp, err := client.UpdateDictionaryDefinitionWithResponse(ctx, current.Id.ValueString(),
			&pim.UpdateDictionaryDefinitionParams{Validation: utils.Ref(pim.UpdateDictionaryDefinitionParamsValidationNAME)},
			toDto(planned),
		)
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update dictionary attribute definition", err.Error())
		}

		if d := utils.AssertSuccessStatusCode(resp); d != nil {
			return nil, d
		}
	}

	if d := updateExtras(ctx, client, current, planned); d != nil {
		return nil, d
	}

	return getExisting(ctx, client, current.Id.ValueString())
}

// updateExtras updates the properties which are not part of the update request:
// the description and the parent definition.
func updateExtras(ctx context.Context, client pim.ClientWithResponsesInterface, current *DictionaryAttributeDefinition, planned *DictionaryAttributeDefinition) diag.Diagnostic {
	id := current.Id.ValueString()

	if !planned.Description.Equal(current.Description) {
		if d := attribute_definition.UpdateDescription(ctx, client, id, planned.Description); d != nil {
			return d
		}
	}

	if !planned.ParentDefinitionId.Equal(current.ParentDefinitionId) {
		if planned.ParentDefinitionId.IsNull() {
			resp, err := client.DeleteDictionaryParentWithResponse(ctx, id)
			if err != nil {
				return diag.NewErrorDiagnostic("Unable to remove parent dictionary", err.Error())
			}

			if d := utils.AssertSuccessStatusCode(resp); d != nil {
				return d
			}
		} else {
			resp, err := client.SetDictionaryParentWithResponse(ctx, id, planned.ParentDefinitionId.ValueString())
			if err != nil {
				return diag.NewErrorDiagnostic("Unable to set parent dictionary", err.Error())
			}

			if d := utils.AssertSuccessStatusCode(resp); d != nil {
				return d
			}
		}
	}

	return nil
}

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*DictionaryAttributeDefinition, diag.Diagnostic) {
	result, d := GetDictionaryAttributeDefinitionByID(ctx, client, id)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Attribute definition not found", "Attribute definition "+id+" was not found after writing it")
	}
	return result, nil
}
