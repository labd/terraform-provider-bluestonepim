package attribute_group_order

import (
	"context"
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

// There is only one order of attribute groups, so the resource has a fixed ID.
const resourceId = "attribute_group_order"

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

type AttributeGroupOrder struct {
	Id       types.String   `tfsdk:"id"`
	GroupIds []types.String `tfsdk:"group_ids"`
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attribute_group_order"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The order of the attribute groups. The listed groups are placed first, in the " +
			"given order; groups that are not listed keep their relative order after them. Removing the " +
			"resource leaves the order as it is. Use only one of these resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Always `" + resourceId + "`.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_ids": schema.ListAttribute{
				MarkdownDescription: "The IDs of the attribute groups, in the order they should be shown.",
				ElementType:         types.StringType,
				Required:            true,
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

func (r *Resource) currentOrder(ctx context.Context) ([]string, diag.Diagnostic) {
	groups, d := attribute_group.ListAttributeGroups(ctx, r.client)
	if d != nil {
		return nil, d
	}
	ids := make([]string, 0, len(groups))
	for _, group := range groups {
		if group.Id != nil {
			ids = append(ids, *group.Id)
		}
	}
	return ids, nil
}

func (r *Resource) apply(ctx context.Context, plan *AttributeGroupOrder) (*AttributeGroupOrder, diag.Diagnostic) {
	current, d := r.currentOrder(ctx)
	if d != nil {
		return nil, d
	}

	desired := make([]string, 0, len(plan.GroupIds))
	for _, id := range plan.GroupIds {
		desired = append(desired, id.ValueString())
	}

	d = utils.Reorder(current, desired, func(id string, position int) diag.Diagnostic {
		response, err := r.client.MoveGroupWithResponse(ctx, id, &pim.MoveGroupParams{TargetPosition: int32(position)})
		if err != nil {
			return diag.NewErrorDiagnostic("Unable to move attribute group", err.Error())
		}
		return utils.AssertStatusCode(response, http.StatusOK)
	})
	if d != nil {
		return nil, d
	}

	return r.read(ctx, len(desired))
}

func (r *Resource) read(ctx context.Context, n int) (*AttributeGroupOrder, diag.Diagnostic) {
	current, d := r.currentOrder(ctx)
	if d != nil {
		return nil, d
	}

	result := &AttributeGroupOrder{Id: types.StringValue(resourceId)}
	for _, id := range utils.LeadingItems(current, n) {
		result.GroupIds = append(result.GroupIds, types.StringValue(id))
	}
	return result, nil
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AttributeGroupOrder
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
	var state AttributeGroupOrder
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	n := len(state.GroupIds)
	if n == 0 {
		// Imported: show all groups
		n = int(^uint(0) >> 1)
	}

	result, diag := r.read(ctx, n)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AttributeGroupOrder
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

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), resourceId)...)
}
