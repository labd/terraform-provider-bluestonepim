package category_attribute

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/bluestonepim-go-sdk/pim"
	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

func fromResponse(r *pim.CategoryAttributeMetadataResponse, categoryId, attributeId string) *CategoryAttribute {
	flags := cla.FlagsFromResponse(r, categoryId)
	result := &CategoryAttribute{
		CategoryId:            types.StringValue(categoryId),
		AttributeDefinitionId: types.StringValue(attributeId),
		Value:                 types.StringNull(),
		Mandatory:             flags.Mandatory,
		Propagate:             flags.Propagate,
		Locked:                flags.Locked,
	}
	// Empty values are stored as null, the schema doesn't allow empty values
	if r.AttributeValue != nil && *r.AttributeValue != "" {
		result.Value = types.StringValue(*r.AttributeValue)
	}
	return result
}

// GetCategoryAttribute returns nil when the attribute is not assigned to the category.
func GetCategoryAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*CategoryAttribute, diag.Diagnostic) {
	r, d := cla.Get(ctx, client, categoryId, attributeId)
	if d != nil || r == nil {
		return nil, d
	}
	return fromResponse(r, categoryId, attributeId), nil
}

func CreateCategoryAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, planned *CategoryAttribute) (*CategoryAttribute, diag.Diagnostic) {
	categoryId := planned.CategoryId.ValueString()
	attributeId := planned.AttributeDefinitionId.ValueString()

	// Assigning a compound attribute also assigns its sub attributes, so the
	// attribute can already be assigned. Take it over instead of failing.
	existing, d := GetCategoryAttribute(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	if existing != nil {
		return UpdateCategoryAttribute(ctx, client, existing, planned)
	}

	response, err := client.CreateCatalogNodeAttributeWithResponse(ctx, categoryId, attributeId,
		&pim.CreateCatalogNodeAttributeParams{ForceCla: utils.Ref(true)},
		pim.CreateCatalogNodeAttributeJSONRequestBody{Value: planned.Value.ValueStringPointer()},
	)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to assign attribute definition to category", err.Error())
	}

	if d := utils.AssertStatusCode(response, http.StatusAccepted); d != nil {
		return nil, d
	}

	// The flags cannot be set on create
	current := *planned
	current.Mandatory = types.BoolValue(false)
	current.Propagate = types.BoolValue(false)
	current.Locked = types.BoolValue(false)
	return UpdateCategoryAttribute(ctx, client, &current, planned)
}

// UpdateCategoryAttribute updates the value and flags in a single request.
func UpdateCategoryAttribute(ctx context.Context, client pim.ClientWithResponsesInterface, current, planned *CategoryAttribute) (*CategoryAttribute, diag.Diagnostic) {
	categoryId := current.CategoryId.ValueString()
	attributeId := current.AttributeDefinitionId.ValueString()

	body := pim.UpdateNodeAttributeValueJSONRequestBody{}
	changed := cla.ApplyChangedFlags(current.flags(), planned.flags(), &body.Mandatory, &body.CopyAttribute, &body.Locked)
	if !planned.Value.Equal(current.Value) {
		body.AttributeValue = &pim.PropertyUpdateString{Value: planned.Value.ValueStringPointer()}
		changed = true
	}

	if changed {
		if d := cla.Patch(ctx, client, categoryId, attributeId, body); d != nil {
			return nil, d
		}
	}

	r, d := cla.MustGet(ctx, client, categoryId, attributeId)
	if d != nil {
		return nil, d
	}
	return fromResponse(r, categoryId, attributeId), nil
}
