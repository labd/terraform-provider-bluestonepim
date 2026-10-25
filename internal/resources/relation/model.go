package relation

import "github.com/hashicorp/terraform-plugin-framework/types"

type Relation struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Number          types.String `tfsdk:"number"`
	Description     types.String `tfsdk:"description"`
	ReverseName     types.String `tfsdk:"reverse_name"`
	Direction       types.String `tfsdk:"direction"`
	QuantityEnabled types.Bool   `tfsdk:"quantity_enabled"`
	ForceDestroy    types.Bool   `tfsdk:"force_destroy"`
}
