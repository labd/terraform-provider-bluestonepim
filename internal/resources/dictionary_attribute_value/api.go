package dictionary_attribute_value

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// defaultContext is the context the API uses when no context is passed.
const defaultContext = "en"

// GetDictionaryAttributeValueByID returns nil when the value does not exist.
func GetDictionaryAttributeValueByID(ctx context.Context, client pim.ClientWithResponsesInterface, definitionId, valueId string) (*DictionaryAttributeValue, diag.Diagnostic) {
	resp, err := client.GetDictionaryAttributeWithResponse(ctx, definitionId, valueId, &pim.GetDictionaryAttributeParams{})
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
	result := &DictionaryAttributeValue{
		Id:                    types.StringValue(valueId),
		AttributeDefinitionId: types.StringValue(definitionId),
		Number:                types.StringPointerValue(resource.Number),
		Value:                 types.StringNull(),
	}
	if resource.Value != nil && resource.Value.Value != nil {
		if v, ok := (*resource.Value.Value)[defaultContext]; ok {
			result.Value = types.StringValue(v)
		}
	}
	return result, nil
}

func CreateDictionaryAttributeValue(ctx context.Context, client pim.ClientWithResponsesInterface, resource *DictionaryAttributeValue) (*DictionaryAttributeValue, diag.Diagnostic) {
	definitionId := resource.AttributeDefinitionId.ValueString()

	resp, err := client.CreateDictionaryAttributeWithResponse(ctx, definitionId, &pim.CreateDictionaryAttributeParams{},
		pim.CreateDictionaryAttributeJSONRequestBody{
			Number: utils.KnownStringPointer(resource.Number),
			Value:  resource.Value.ValueString(),
		})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create dictionary value", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")
	return getExisting(ctx, client, definitionId, resourceId)
}

func UpdateDictionaryAttributeValue(ctx context.Context, client pim.ClientWithResponsesInterface, current *DictionaryAttributeValue, planned *DictionaryAttributeValue) (*DictionaryAttributeValue, diag.Diagnostic) {
	body := pim.UpdateDictionaryAttributeJSONRequestBody{}
	if !planned.Value.Equal(current.Value) {
		body.Value = &pim.PropertyUpdateString{Value: planned.Value.ValueStringPointer()}
	}
	if !planned.Number.IsUnknown() && !planned.Number.Equal(current.Number) {
		body.Number = &pim.PropertyUpdateString{Value: planned.Number.ValueStringPointer()}
	}

	if body.Value != nil || body.Number != nil {
		resp, err := client.UpdateDictionaryAttributeWithResponse(ctx,
			current.AttributeDefinitionId.ValueString(), current.Id.ValueString(), &pim.UpdateDictionaryAttributeParams{}, body)
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update dictionary value", err.Error())
		}

		if d := utils.AssertSuccessStatusCode(resp); d != nil {
			return nil, d
		}
	}

	return getExisting(ctx, client, current.AttributeDefinitionId.ValueString(), current.Id.ValueString())
}

func DeleteDictionaryAttributeValue(ctx context.Context, client pim.ClientWithResponsesInterface, definitionId, valueId string) diag.Diagnostic {
	resp, err := client.DeleteDictionaryAttributeWithResponse(ctx, definitionId, valueId)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete dictionary value", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil
	}

	return utils.AssertSuccessStatusCode(resp)
}

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, definitionId, valueId string) (*DictionaryAttributeValue, diag.Diagnostic) {
	result, d := GetDictionaryAttributeValueByID(ctx, client, definitionId, valueId)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Dictionary value not found", "Dictionary value "+valueId+" was not found after writing it")
	}
	return result, nil
}
