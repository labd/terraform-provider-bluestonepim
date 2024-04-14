package category_attribute

import "github.com/hashicorp/terraform-plugin-framework/types"

type CategoryAttribute struct {
	CategoryId  types.String `tfsdk:"category_id"`
	AttributeId types.String `tfsdk:"attribute_id"`
	Mandatory   types.Bool   `tfsdk:"mandatory"`
}
