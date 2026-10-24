package category_context

import "github.com/hashicorp/terraform-plugin-framework/types"

type CategoryContext struct {
	CategoryId  types.String `tfsdk:"category_id"`
	ContextId   types.String `tfsdk:"context_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}
