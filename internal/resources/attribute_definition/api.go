package attribute_definition

import (
	"context"
	"github.com/labd/bluestonepim-go-sdk/pim"
	"net/http"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func GetAttributeDefinitionByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*AttributeDefinition, diag.Diagnostic) {
	resp, err := client.GetAttributeDefinitionWithResponse(ctx, id, nil)

	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	resource := resp.JSON200

	result := &AttributeDefinition{
		Id:             types.StringPointerValue(resource.Id),
		Name:           types.StringValue(resource.Name),
		Description:    types.StringPointerValue(resource.Description),
		Number:         types.StringPointerValue(resource.Number),
		DataType:       types.StringValue(string(*resource.DataType)),
		CharacterSet:   types.StringPointerValue(resource.Charset),
		ContentType:    types.StringPointerValue(resource.ContentType),
		ExternalSource: types.BoolPointerValue(resource.ExternalSource),
		Internal:       types.BoolPointerValue(resource.Internal),
		GroupID:        types.StringPointerValue(resource.GroupId),
		Unit:           types.StringPointerValue(resource.Unit),
		Restrictions:   FromRestrictionsDto(resource.Restrictions),
	}
	return result, nil
}

func FromRestrictionsDto(restrictions *pim.RestrictionsDto) *Restrictions {
	if restrictions == nil {
		return nil
	}

	var dto = &Restrictions{}
	if restrictions.Enum != nil {
		values := make([]EnumValue, 0, len(*restrictions.Enum.Values))
		for _, v := range *restrictions.Enum.Values {
			values = append(values, EnumValue{
				Metadata: types.StringPointerValue(v.Metadata),
				Number:   types.StringPointerValue(v.Number),
				Value:    types.StringValue(v.Value),
				ValueId:  types.StringPointerValue(v.ValueId),
			})
		}

		dto.Enum = &EnumRestriction{
			Type:   types.StringPointerValue(restrictions.Enum.Type),
			Values: &values,
		}
	}
	if restrictions.Range != nil {
		dto.Range = &RangeRestriction{
			Max:  types.StringPointerValue(restrictions.Range.Max),
			Min:  types.StringPointerValue(restrictions.Range.Min),
			Step: types.StringPointerValue(restrictions.Range.Step),
		}

	}

	if restrictions.Text != nil {
		dto.Text = &TextRestriction{
			MaxLength:   types.Int32PointerValue(restrictions.Text.MaxLength),
			Pattern:     types.StringPointerValue(restrictions.Text.Pattern),
			Whitespaces: types.BoolPointerValue(restrictions.Text.Whitespaces),
		}

	}

	return dto
}

func CreateAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, resource *AttributeDefinition) (*AttributeDefinition, diag.Diagnostic) {
	resC, err := client.CreateAttributeDefinitionWithResponse(ctx,
		&pim.CreateAttributeDefinitionParams{
			Validation: utils.Ref[pim.CreateAttributeDefinitionParamsValidation]("NAME"),
		},
		pim.CreateAttributeDefinitionJSONRequestBody{
			Charset:        resource.CharacterSet.ValueStringPointer(),
			ContentType:    resource.ContentType.ValueStringPointer(),
			DataType:       utils.Ref(pim.SimpleAttributeDefinitionRequestDataType(resource.DataType.ValueString())),
			ExternalSource: resource.ExternalSource.ValueBoolPointer(),
			Internal:       resource.Internal.ValueBoolPointer(),
			GroupId:        resource.GroupID.ValueStringPointer(),
			Name:           resource.Name.ValueString(),
			Number:         resource.Number.ValueStringPointer(),
			Unit:           resource.Unit.ValueStringPointer(),
			Restrictions:   ToRestrictionsDto(resource.Restrictions),
		},
	)
	if err != nil {
		d := diag.NewErrorDiagnostic("Unable to create attribute definition", err.Error())
		return nil, d
	}

	if d := utils.AssertStatusCode(resC, http.StatusCreated); d != nil {
		return nil, d
	}

	resourceId := resC.HTTPResponse.Header.Get("Resource-Id")

	//Workaround because we cannot set description on create
	resU, err := client.UpdateMetadataWithResponse(ctx, resourceId, nil,
		pim.UpdateMetadataJSONRequestBody{
			Description: &pim.PropertyUpdateString{
				Value: resource.Description.ValueStringPointer(),
			},
		})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to update attribute definition description", err.Error())
	}
	if d := utils.AssertStatusCode(resU, http.StatusNoContent); d != nil {
		return nil, d
	}

	return GetAttributeDefinitionByID(ctx, client, resourceId)
}

// Does not include description as this needs to be updated through the metadata
func attributeDefinitionHasChanges(current *AttributeDefinition, planned *AttributeDefinition) bool {
	if !reflect.DeepEqual(current.Restrictions, planned.Restrictions) {
		return true
	}

	return !(planned.Name.Equal(current.Name) &&
		planned.Number.Equal(current.Number) &&
		planned.DataType.Equal(current.DataType) &&
		planned.ContentType.Equal(current.ContentType) &&
		planned.CharacterSet.Equal(current.CharacterSet) &&
		planned.ExternalSource.Equal(current.ExternalSource) &&
		planned.Internal.Equal(current.Internal) &&
		planned.GroupID.Equal(current.GroupID) &&
		planned.Unit.Equal(current.Unit))

}

func UpdateAttributeDefinition(ctx context.Context, client pim.ClientWithResponsesInterface, current *AttributeDefinition, planned *AttributeDefinition) (*AttributeDefinition, diag.Diagnostic) {
	if attributeDefinitionHasChanges(current, planned) {
		res, err := client.UpdateAttributeDefinitionWithResponse(ctx, current.Id.ValueString(), nil,
			pim.UpdateAttributeDefinitionJSONRequestBody{
				Charset:        planned.CharacterSet.ValueStringPointer(),
				ContentType:    planned.ContentType.ValueStringPointer(),
				DataType:       utils.Ref(pim.SimpleAttributeDefinitionRequestDataType(planned.DataType.ValueString())),
				ExternalSource: planned.ExternalSource.ValueBoolPointer(),
				Internal:       planned.Internal.ValueBoolPointer(),
				GroupId:        planned.GroupID.ValueStringPointer(),
				Name:           planned.Name.ValueString(),
				Number:         planned.Number.ValueStringPointer(),
				Unit:           planned.Unit.ValueStringPointer(),
				Restrictions:   ToRestrictionsDto(planned.Restrictions),
			})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update attribute definition", err.Error())
		}

		if d := utils.AssertStatusCode(res, http.StatusOK); d != nil {
			return nil, d
		}
	}

	if !planned.Description.Equal(current.Description) {
		//Workaround because we cannot set description on create
		resU, err := client.UpdateMetadataWithResponse(ctx, current.Id.ValueString(), nil,
			pim.UpdateMetadataJSONRequestBody{
				Description: &pim.PropertyUpdateString{
					Value: planned.Description.ValueStringPointer(),
				},
			})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update attribute definition description", err.Error())
		}
		if d := utils.AssertStatusCode(resU, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	return GetAttributeDefinitionByID(ctx, client, current.Id.ValueString())
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
