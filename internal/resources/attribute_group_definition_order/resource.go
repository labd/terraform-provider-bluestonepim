package attribute_group_definition_order

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_group"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

const pageSize = 1000

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client pim.ClientWithResponsesInterface
}

type DefinitionOrder struct {
	GroupId       types.String   `tfsdk:"group_id"`
	DefinitionIds []types.String `tfsdk:"definition_ids"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attribute_group_definition_order"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The manual order of the attribute definitions in an attribute group. The group " +
			"must have `order_by_name = false`. The listed definitions are placed first, in the given order; " +
			"definitions that are not listed keep their relative order after them. Removing the resource " +
			"leaves the order as it is.",
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the attribute group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"definition_ids": schema.ListAttribute{
				MarkdownDescription: "The IDs of the attribute definitions in the group, in the order they should " +
					"be shown. The definitions must belong to the group (see `group_id` on the definitions).",
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.UniqueValues(),
				},
			},
		},
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	data, diag := utils.GetProviderData(req.ProviderData)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	r.client = data.PimClient
}

// currentOrder returns the definitions in the group in the order the API
// returns them (which is assumed to be the display order), or nil when the
// group does not exist.
func (r *Resource) currentOrder(ctx context.Context, groupId string) ([]string, diag.Diagnostic) {
	ids := []string{}
	for page := int32(0); ; page++ {
		response, err := r.client.FindDefinitionsInGroupWithResponse(ctx, groupId, &pim.FindDefinitionsInGroupParams{
			Page:     utils.Ref(page),
			PageSize: utils.Ref(int32(pageSize)),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
		}

		if utils.IsNotFound(response) {
			return nil, nil
		}

		if d := utils.AssertStatusCode(response, http.StatusOK); d != nil {
			return nil, d
		}

		if response.JSON200.Data == nil {
			return ids, nil
		}
		for _, definition := range *response.JSON200.Data {
			if definition.Id != nil {
				ids = append(ids, *definition.Id)
			}
		}
		if len(*response.JSON200.Data) < pageSize {
			return ids, nil
		}
	}
}

func (r *Resource) apply(ctx context.Context, plan *DefinitionOrder) (*DefinitionOrder, diag.Diagnostic) {
	groupId := plan.GroupId.ValueString()

	// A manual order only applies when the group isn't ordered by name. The
	// order strategy is managed on the group, so don't change it here.
	group, d := attribute_group.GetAttributeGroupByID(ctx, r.client, groupId)
	if d != nil {
		return nil, d
	}
	if group == nil {
		return nil, utils.NotFoundError("Attribute group", groupId)
	}
	if group.OrderByName.ValueBool() {
		return nil, diag.NewErrorDiagnostic("Attribute group is ordered by name",
			fmt.Sprintf("The definitions in attribute group %s are ordered by name. Set `order_by_name = false` on "+
				"the bluestonepim_attribute_group to order them manually.", groupId))
	}

	current, d := r.currentOrder(ctx, groupId)
	if d != nil {
		return nil, d
	}

	desired := make([]string, 0, len(plan.DefinitionIds))
	for _, id := range plan.DefinitionIds {
		desired = append(desired, id.ValueString())
	}

	d = utils.Reorder(current, desired, func(id string, position int) diag.Diagnostic {
		response, err := r.client.MoveAttributeDefinitionWithResponse(ctx, groupId, id,
			&pim.MoveAttributeDefinitionParams{TargetPosition: int32(position)})
		if err != nil {
			return diag.NewErrorDiagnostic("Unable to move attribute definition", err.Error())
		}
		return utils.AssertStatusCode(response, http.StatusOK)
	})
	if d != nil {
		return nil, diag.NewErrorDiagnostic(d.Summary(),
			d.Detail()+fmt.Sprintf(" in attribute group %s. Check that the definition has this group as `group_id`.", groupId))
	}

	result, d := r.read(ctx, groupId, len(desired))
	if d == nil && result == nil {
		d = utils.NotFoundError("Attribute group", groupId)
	}
	return result, d
}

func (r *Resource) read(ctx context.Context, groupId string, n int) (*DefinitionOrder, diag.Diagnostic) {
	current, d := r.currentOrder(ctx, groupId)
	if d != nil || current == nil {
		return nil, d
	}

	result := &DefinitionOrder{GroupId: types.StringValue(groupId)}
	for _, id := range utils.LeadingItems(current, n) {
		result.DefinitionIds = append(result.DefinitionIds, types.StringValue(id))
	}
	return result, nil
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DefinitionOrder
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := r.apply(ctx, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DefinitionOrder
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	n := len(state.DefinitionIds)
	if n == 0 {
		// Imported: show all definitions in the group
		n = int(^uint(0) >> 1)
	}

	result, diag := r.read(ctx, state.GroupId.ValueString(), n)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DefinitionOrder
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := r.apply(ctx, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Delete leaves the order as it is.
func (r *Resource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}

// ImportState imports the order of a group by its ID.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("group_id"), req, resp)
}
