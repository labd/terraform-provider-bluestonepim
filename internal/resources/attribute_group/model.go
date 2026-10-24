package attribute_group

import "github.com/hashicorp/terraform-plugin-framework/types"

type AttributeGroup struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Number types.String `tfsdk:"number"`

	OrderByName types.Bool `tfsdk:"order_by_name"`
}
