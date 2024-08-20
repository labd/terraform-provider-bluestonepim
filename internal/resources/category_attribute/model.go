package category_attribute

import "github.com/hashicorp/terraform-plugin-framework/types"

type CategoryAttribute struct {
	CategoryId            types.String `tfsdk:"category_id"`
	AttributeDefinitionId types.String `tfsdk:"attribute_definition_id"`
	Mandatory             types.Bool   `tfsdk:"mandatory"`
}
