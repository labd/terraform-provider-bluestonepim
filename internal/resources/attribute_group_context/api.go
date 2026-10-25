package attribute_group_context

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_group"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

type GroupContext struct {
	GroupId   types.String `tfsdk:"group_id"`
	ContextId types.String `tfsdk:"context_id"`
	Name      types.String `tfsdk:"name"`
}

// GetGroupContext reads the name stored for the context, without falling back
// to the name of other contexts. Returns nil when the group does not exist.
func GetGroupContext(ctx context.Context, client pim.ClientWithResponsesInterface, groupId, contextId string) (*GroupContext, diag.Diagnostic) {
	groups, d := attribute_group.ListAttributeGroupsInContext(ctx, client, utils.Ref(contextId))
	if d != nil {
		return nil, d
	}

	for _, group := range groups {
		if group.Id == nil || *group.Id != groupId {
			continue
		}
		result := &GroupContext{
			GroupId:   types.StringValue(groupId),
			ContextId: types.StringValue(contextId),
			Name:      types.StringNull(),
		}
		if group.Name != "" {
			result.Name = types.StringValue(group.Name)
		}
		return result, nil
	}
	return nil, nil
}

// SetGroupName writes the name for the context.
func SetGroupName(ctx context.Context, client pim.ClientWithResponsesInterface, groupId, contextId, name string) diag.Diagnostic {
	resp, err := client.RenameWithResponse(ctx, groupId, &pim.RenameParams{Context: utils.Ref(contextId)},
		pim.RenameJSONRequestBody{Name: &name})
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to rename attribute group", err.Error())
	}
	return utils.AssertStatusCode(resp, http.StatusNoContent)
}

func mustGet(ctx context.Context, client pim.ClientWithResponsesInterface, groupId, contextId string) (*GroupContext, diag.Diagnostic) {
	result, d := GetGroupContext(ctx, client, groupId, contextId)
	if d == nil && result == nil {
		d = utils.NotFoundError("Attribute group", groupId)
	}
	return result, d
}
