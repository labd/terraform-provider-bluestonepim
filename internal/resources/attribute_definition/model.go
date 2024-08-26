package attribute_definition

import "github.com/hashicorp/terraform-plugin-framework/types"

type AttributeDefinition struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Number         types.String `tfsdk:"number"`
	DataType       types.String `tfsdk:"data_type"`
	ContentType    types.String `tfsdk:"content_type"`
	CharacterSet   types.String `tfsdk:"character_set"`
	ExternalSource types.Bool   `tfsdk:"external_source"`
	GroupID        types.String `tfsdk:"group_id"`
	Internal       types.Bool   `tfsdk:"internal"`
	Unit           types.String `tfsdk:"unit"`
}
