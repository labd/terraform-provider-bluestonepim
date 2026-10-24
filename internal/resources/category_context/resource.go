package category_context

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &Resource{}
	_ resource.ResourceWithConfigure        = &Resource{}
	_ resource.ResourceWithImportState      = &Resource{}
	_ resource.ResourceWithConfigValidators = &Resource{}
)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client pim.ClientWithResponsesInterface
}

// Metadata returns the data source type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_category_context"
}

// Schema defines the schema for the data source.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sets the name and / or description of a category for a specific context, for " +
			"example another language or a sales channel. Values that are not set fall back to the fallback " +
			"context. The category tree itself is the same for every context.",
		Attributes: map[string]schema.Attribute{
			"category_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the category.",
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
				MarkdownDescription: "The name of the category in this context.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the category in this context.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (r *Resource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(path.MatchRoot("name"), path.MatchRoot("description")),
	}
}

// Configure adds the provider configured client to the data source.
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

// Create creates the resource and sets the initial Terraform state.
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CategoryContext
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Compare against the current values, so only the differences are written
	current, diag := GetCategoryContext(ctx, r.client, plan.CategoryId.ValueString(), plan.ContextId.ValueString())
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
	if current == nil {
		resp.Diagnostics.AddError("Category not found", "Category "+plan.CategoryId.ValueString()+" was not found")
		return
	}

	result, diag := SetCategoryContext(ctx, r.client, current, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var current CategoryContext
	diags := req.State.Get(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := GetCategoryContext(ctx, r.client, current.CategoryId.ValueString(), current.ContextId.ValueString())
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CategoryContext
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CategoryContext
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := SetCategoryContext(ctx, r.client, &state, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
}

// Delete clears the values of the context, so they fall back to the fallback
// context again.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CategoryContext
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cleared := state
	cleared.Name = types.StringNull()
	cleared.Description = types.StringNull()

	_, diag := SetCategoryContext(ctx, r.client, &state, &cleared)
	if diag != nil {
		resp.Diagnostics.Append(diag)
	}
}

// ImportState imports the context values by `<category_id>/<context_id>`.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected `<category_id>/<context_id>`, got %q", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("category_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("context_id"), parts[1])...)
}
