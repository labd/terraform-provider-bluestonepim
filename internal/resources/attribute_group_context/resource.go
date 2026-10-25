package attribute_group_context

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

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

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_attribute_group_context"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets the name of an attribute group for a specific context, for example another " +
			"language or a sales channel.\n\n" +
			"The API has no way to remove the name of a context, so removing this resource leaves the name in " +
			"Bluestone as it is.",
		Attributes: map[string]schema.Attribute{
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the attribute group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"context_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the context (see `bluestonepim_context`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the attribute group in this context.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
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

func (r *Resource) set(ctx context.Context, current, plan *GroupContext, diags *diag.Diagnostics) *GroupContext {
	groupId := plan.GroupId.ValueString()
	contextId := plan.ContextId.ValueString()

	if current == nil || !plan.Name.Equal(current.Name) {
		if d := SetGroupName(ctx, r.client, groupId, contextId, plan.Name.ValueString()); d != nil {
			diags.Append(d)
			return nil
		}
	}

	result, d := mustGet(ctx, r.client, groupId, contextId)
	if d != nil {
		diags.Append(d)
		return nil
	}
	return result
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupContext
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	current, d := GetGroupContext(ctx, r.client, plan.GroupId.ValueString(), plan.ContextId.ValueString())
	if d != nil {
		resp.Diagnostics.Append(d)
		return
	}
	if current == nil {
		resp.Diagnostics.Append(utils.NotFoundError("Attribute group", plan.GroupId.ValueString()))
		return
	}

	result := r.set(ctx, current, &plan, &resp.Diagnostics)
	if result != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	}
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GroupContext
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, d := GetGroupContext(ctx, r.client, state.GroupId.ValueString(), state.ContextId.ValueString())
	if d != nil {
		resp.Diagnostics.Append(d)
		return
	}

	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state GroupContext
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result := r.set(ctx, &state, &plan, &resp.Diagnostics)
	if result != nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
	}
}

// Delete leaves the name in Bluestone, the API cannot remove the name of a
// context.
func (r *Resource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {}

// ImportState imports the name by `<group_id>/<context_id>`.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected `<group_id>/<context_id>`, got %q", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("context_id"), parts[1])...)
}
