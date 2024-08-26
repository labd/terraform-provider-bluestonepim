package category

import "github.com/hashicorp/terraform-plugin-framework/types"

// Category describes the data source data model.
type Category struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Number      types.String `tfsdk:"number"`
	Description types.String `tfsdk:"description"`
	ContextId   types.String `tfsdk:"context_id"`
	ParentId    types.String `tfsdk:"parent_id"`
}
