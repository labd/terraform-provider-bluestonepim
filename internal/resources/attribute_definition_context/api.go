package attribute_definition_context

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

type DefinitionContext struct {
	DefinitionId types.String `tfsdk:"definition_id"`
	ContextId    types.String `tfsdk:"context_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
}

// GetDefinitionContext reads the values stored for the context, without
// falling back to the values of other contexts. Returns nil when the
// definition does not exist.
func GetDefinitionContext(ctx context.Context, client pim.ClientWithResponsesInterface, definitionId, contextId string) (*DefinitionContext, diag.Diagnostic) {
	resp, err := client.GetAttributeDefinitionWithResponse(ctx, definitionId, &pim.GetAttributeDefinitionParams{
		Context:         utils.Ref(contextId),
		ContextFallback: utils.Ref(false),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	return &DefinitionContext{
		DefinitionId: types.StringValue(definitionId),
		ContextId:    types.StringValue(contextId),
		Name:         nullIfEmpty(resp.JSON200.Name),
		Description:  nullIfEmpty(resp.JSON200.Description),
	}, nil
}

// SetDefinitionContext writes the changed values for the context. A null value
// clears the value in the context, so it falls back to the fallback context.
// The metadata endpoint works for every definition type.
func SetDefinitionContext(ctx context.Context, client pim.ClientWithResponsesInterface, current, planned *DefinitionContext) (*DefinitionContext, diag.Diagnostic) {
	definitionId := planned.DefinitionId.ValueString()
	contextId := planned.ContextId.ValueString()

	body := pim.UpdateMetadataJSONRequestBody{}
	if !planned.Name.Equal(current.Name) {
		body.Name = &pim.PropertyUpdateString{Value: planned.Name.ValueStringPointer()}
	}
	if !planned.Description.Equal(current.Description) {
		body.Description = &pim.PropertyUpdateString{Value: planned.Description.ValueStringPointer()}
	}

	if body.Name != nil || body.Description != nil {
		resp, err := client.UpdateMetadataWithResponse(ctx, definitionId,
			&pim.UpdateMetadataParams{Context: utils.Ref(contextId)}, body)
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update attribute definition", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	result, d := GetDefinitionContext(ctx, client, definitionId, contextId)
	if d == nil && result == nil {
		d = utils.NotFoundError("Attribute definition", definitionId)
	}
	return result, d
}

func nullIfEmpty(v *string) types.String {
	if v == nil || *v == "" {
		return types.StringNull()
	}
	return types.StringValue(*v)
}
