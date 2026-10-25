package attribute_group

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

const pageSize = 1000

// ListAttributeGroups returns all attribute groups, in the order the API
// returns them (which is assumed to be the display order).
func ListAttributeGroups(ctx context.Context, client pim.ClientWithResponsesInterface) ([]pim.AttributeGroupResponse, diag.Diagnostic) {
	return ListAttributeGroupsInContext(ctx, client, nil)
}

// ListAttributeGroupsInContext returns all attribute groups with the names of
// the given context, without falling back to other contexts. A nil context
// uses the default context.
func ListAttributeGroupsInContext(ctx context.Context, client pim.ClientWithResponsesInterface, contextId *string) ([]pim.AttributeGroupResponse, diag.Diagnostic) {
	var fallback *bool
	if contextId != nil {
		fallback = utils.Ref(false)
	}

	var result []pim.AttributeGroupResponse
	for page := int32(0); ; page++ {
		resp, err := client.FindAttributeGroupsWithResponse(ctx, &pim.FindAttributeGroupsParams{
			Page:            utils.Ref(page),
			PageSize:        utils.Ref(int32(pageSize)),
			Context:         contextId,
			ContextFallback: fallback,
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
			return nil, d
		}

		if resp.JSON200.Data == nil {
			return result, nil
		}

		result = append(result, *resp.JSON200.Data...)
		if len(*resp.JSON200.Data) < pageSize {
			return result, nil
		}
	}
}

// GetAttributeGroupByID returns nil when the group does not exist. The API has
// no endpoint to fetch a single group, so we list all groups.
func GetAttributeGroupByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*AttributeGroup, diag.Diagnostic) {
	groups, d := ListAttributeGroups(ctx, client)
	if d != nil {
		return nil, d
	}

	for _, group := range groups {
		if group.Id != nil && *group.Id == id {
			return &AttributeGroup{
				Id:          types.StringPointerValue(group.Id),
				Name:        types.StringValue(group.Name),
				Number:      types.StringPointerValue(group.Number),
				OrderByName: types.BoolPointerValue(group.OrderByName),
			}, nil
		}
	}
	return nil, nil
}

func CreateAttributeGroup(ctx context.Context, client pim.ClientWithResponsesInterface, resource *AttributeGroup) (*AttributeGroup, diag.Diagnostic) {
	resp, err := client.CreateAttributeGroupWithResponse(ctx, &pim.CreateAttributeGroupParams{}, pim.CreateAttributeGroupJSONRequestBody{
		Name:   resource.Name.ValueString(),
		Number: utils.KnownStringPointer(resource.Number),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create attribute group", err.Error())
	}

	if d := utils.AssertStatusCode(resp, http.StatusCreated); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")

	// The order strategy cannot be set on create
	if isKnown(resource.OrderByName) {
		if d := setOrderByName(ctx, client, resourceId, resource.OrderByName.ValueBool()); d != nil {
			return nil, d
		}
	}

	return getExisting(ctx, client, resourceId)
}

func UpdateAttributeGroup(ctx context.Context, client pim.ClientWithResponsesInterface, current *AttributeGroup, planned *AttributeGroup) (*AttributeGroup, diag.Diagnostic) {
	id := current.Id.ValueString()

	if !planned.Name.Equal(current.Name) {
		resp, err := client.RenameWithResponse(ctx, id, &pim.RenameParams{}, pim.RenameJSONRequestBody{
			Name: planned.Name.ValueStringPointer(),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to rename attribute group", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	if isKnown(planned.Number) && !planned.Number.Equal(current.Number) {
		resp, err := client.UpdateNumberWithResponse(ctx, id, pim.UpdateNumberJSONRequestBody{
			Number: planned.Number.ValueString(),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update attribute group number", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	if isKnown(planned.OrderByName) && !planned.OrderByName.Equal(current.OrderByName) {
		if d := setOrderByName(ctx, client, id, planned.OrderByName.ValueBool()); d != nil {
			return nil, d
		}
	}

	return getExisting(ctx, client, id)
}

func setOrderByName(ctx context.Context, client pim.ClientWithResponsesInterface, id string, orderByName bool) diag.Diagnostic {
	resp, err := client.SetDefinitionsOrderStrategyInGroupWithResponse(ctx, id,
		&pim.SetDefinitionsOrderStrategyInGroupParams{OrderByName: orderByName})
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to update attribute group order strategy", err.Error())
	}
	return utils.AssertStatusCode(resp, http.StatusOK)
}

func DeleteAttributeGroup(ctx context.Context, client pim.ClientWithResponsesInterface, id string) diag.Diagnostic {
	resp, err := client.DeleteAttributeGroupWithResponse(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete attribute group", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil
	}

	return utils.AssertStatusCode(resp, http.StatusNoContent)
}

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*AttributeGroup, diag.Diagnostic) {
	result, d := GetAttributeGroupByID(ctx, client, id)
	if d == nil && result == nil {
		d = utils.NotFoundError("Attribute group", id)
	}
	return result, d
}

func isKnown(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}
