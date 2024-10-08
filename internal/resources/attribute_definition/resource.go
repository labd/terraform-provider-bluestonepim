package attribute_definition

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

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
	resp.TypeName = req.ProviderTypeName + "_attribute_definition"
}

// Schema defines the schema for the data source.
func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Platform-generated unique identifier of the Category.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"number": schema.StringAttribute{
				MarkdownDescription: "Number",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Category.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the attribute.",
				Optional:            true,
			},

			"data_type": schema.StringAttribute{
				MarkdownDescription: "The data type of the attribute. For the `matrix`, `dictionary`, and `column` " +
					"data types use the dedicated resources",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"boolean", "integer", "decimal", "date", "time",
						"date_time", "location", "single_select", "multi_select",
						"text", "formatted_text", "pattern", "multiline",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type": schema.StringAttribute{
				MarkdownDescription: "The content type of the attribute.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("text/markdown", "html"),
				},
				Default: stringdefault.StaticString("text/markdown"),
			},
			"character_set": schema.StringAttribute{
				MarkdownDescription: "The unit of the attribute.",
				Optional:            true,
			},
			"external_source": schema.BoolAttribute{
				MarkdownDescription: "Whether the attribute is an external source.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Whether the attribute is internal.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"group_id": schema.StringAttribute{
				MarkdownDescription: "The group ID of the attribute.",
				Optional:            true,
			},
			"unit": schema.StringAttribute{
				MarkdownDescription: "The unit of the attribute.",
				Optional:            true,
			},
			"restrictions": schema.SingleNestedAttribute{
				MarkdownDescription: "The restrictions of the attribute.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"enum": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: "The type of the enum.",
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("text"),
								Validators: []validator.String{
									stringvalidator.OneOf("text", "color"),
								},
							},
							"values": schema.ListNestedAttribute{
								Optional: true,
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"metadata": schema.StringAttribute{
											MarkdownDescription: "The metadata of the enum.",
											Optional:            true,
										},
										"number": schema.StringAttribute{
											MarkdownDescription: "The number of the enum.",
											Optional:            true,
										},
										"value": schema.StringAttribute{
											MarkdownDescription: "The value of the enum.",
											Required:            true,
										},
										"value_id": schema.StringAttribute{
											MarkdownDescription: "The ID of the value.",
											Computed:            true,
										},
									},
								},
							},
						},
					},
					"range": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"max": schema.StringAttribute{
								MarkdownDescription: "The maximum value of the range.",
								Optional:            true,
							},
							"min": schema.StringAttribute{
								MarkdownDescription: "The minimum value of the range.",
								Optional:            true,
							},
							"step": schema.StringAttribute{
								MarkdownDescription: "The step value of the range.",
								Optional:            true,
							},
						},
					},
					"text": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"max_length": schema.Int32Attribute{
								MarkdownDescription: "The maximum length of the text.",
								Optional:            true,
							},
							"pattern": schema.StringAttribute{
								MarkdownDescription: "The pattern of the text.",
								Optional:            true,
							},
							"whitespaces": schema.BoolAttribute{
								MarkdownDescription: "Whether the text allows whitespaces.",
								Optional:            true,
							},
						},
					},
				},
			},
		},
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
	var plan AttributeDefinition
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := CreateAttributeDefinition(ctx, r.client, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var current AttributeDefinition
	diags := req.State.Get(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := GetAttributeDefinitionByID(ctx, r.client, current.Id.ValueString())
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan AttributeDefinition
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state AttributeDefinition
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, diag := UpdateAttributeDefinition(ctx, r.client, &state, &plan)
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state AttributeDefinition
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diag := DeleteAttributeDefinition(ctx, r.client, state.Id.ValueString())
	if diag != nil {
		resp.Diagnostics.Append(diag)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
