package context

import "github.com/hashicorp/terraform-plugin-framework/types"

type Context struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Locale     types.String `tfsdk:"locale"`
	FallbackID types.String `tfsdk:"fallback_id"`
}
