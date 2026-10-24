package attribute_group

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

const pageSize = 1000

// GetAttributeGroupByID returns nil when the group does not exist. The API has
// no endpoint to fetch a single group, so we page through all groups.
func GetAttributeGroupByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*AttributeGroup, diag.Diagnostic) {
	for page := int32(0); ; page++ {
		resp, err := client.FindAttributeGroupsWithResponse(ctx, &pim.FindAttributeGroupsParams{
			Page:     utils.Ref(page),
			PageSize: utils.Ref(int32(pageSize)),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
			return nil, d
		}

		if resp.JSON200.Data == nil {
			return nil, nil
		}

		for _, group := range *resp.JSON200.Data {
			if group.Id != nil && *group.Id == id {
				return &AttributeGroup{
					Id:     types.StringPointerValue(group.Id),
					Name:   types.StringValue(group.Name),
					Number: types.StringPointerValue(group.Number),
				}, nil
			}
		}

		if len(*resp.JSON200.Data) < pageSize {
			return nil, nil
		}
	}
}

func CreateAttributeGroup(ctx context.Context, client pim.ClientWithResponsesInterface, resource *AttributeGroup) (*AttributeGroup, diag.Diagnostic) {
	resp, err := client.CreateAttributeGroupWithResponse(ctx, &pim.CreateAttributeGroupParams{}, pim.CreateAttributeGroupJSONRequestBody{
		Name:   resource.Name.ValueString(),
		Number: utils.KnownStringPointer(resource.Number),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create attribute group", err.Error())
	}

	if d := utils.AssertSuccessStatusCode(resp); d != nil {
		return nil, d
	}

	resourceId := resp.HTTPResponse.Header.Get("Resource-Id")
	return getExisting(ctx, client, resourceId)
}

func UpdateAttributeGroup(ctx context.Context, client pim.ClientWithResponsesInterface, current *AttributeGroup, planned *AttributeGroup) (*AttributeGroup, diag.Diagnostic) {
	if !planned.Name.Equal(current.Name) {
		resp, err := client.RenameWithResponse(ctx, current.Id.ValueString(), &pim.RenameParams{}, pim.RenameJSONRequestBody{
			Name: planned.Name.ValueStringPointer(),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to rename attribute group", err.Error())
		}

		if d := utils.AssertSuccessStatusCode(resp); d != nil {
			return nil, d
		}
	}

	return getExisting(ctx, client, current.Id.ValueString())
}

func DeleteAttributeGroup(ctx context.Context, client pim.ClientWithResponsesInterface, id string) diag.Diagnostic {
	resp, err := client.DeleteAttributeGroupWithResponse(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete attribute group", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil
	}

	return utils.AssertSuccessStatusCode(resp)
}

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*AttributeGroup, diag.Diagnostic) {
	result, d := GetAttributeGroupByID(ctx, client, id)
	if d != nil {
		return nil, d
	}
	if result == nil {
		return nil, diag.NewErrorDiagnostic("Attribute group not found", "Attribute group "+id+" was not found after writing it")
	}
	return result, nil
}
