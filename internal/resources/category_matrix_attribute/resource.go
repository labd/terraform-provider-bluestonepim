package category_matrix_attribute

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
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

// Metadata returns the data source type name.
func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_category_matrix_attribute"
}

// Schema defines the schema for the data source.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := cla.SchemaAttributes(false)
	attributes["values"] = schema.MapAttribute{
		MarkdownDescription: "The value per cell, as a map of column ID to a map of row ID to value (see " +
			"`columns` and `rows` of `bluestonepim_matrix_attribute_definition`). Cells left out are empty.",
		ElementType: valuesType,
		Optional:    true,
		Validators: []validator.Map{
			mapvalidator.SizeAtLeast(1),
		},
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Assigns a `matrix` attribute definition to a category (Category Level Attribute).",
		Attributes:          attributes,
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
	var plan CategoryMatrixAttribute
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := CreateCategoryMatrixAttribute(ctx, r.client, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var current CategoryMatrixAttribute
	resp.Diagnostics.Append(req.State.Get(ctx, &current)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := GetCategoryMatrixAttribute(
		ctx, r.client, current.CategoryId.ValueString(), current.AttributeDefinitionId.ValueString())
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

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state CategoryMatrixAttribute
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := UpdateCategoryMatrixAttribute(ctx, r.client, &state, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, result)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CategoryMatrixAttribute
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diag := cla.Delete(ctx, r.client, state.CategoryId.ValueString(), state.AttributeDefinitionId.ValueString()); diag != nil {
		resp.Diagnostics.Append(diag)
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	cla.ImportState(ctx, req, resp)
}
