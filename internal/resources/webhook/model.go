package webhook

import "github.com/hashicorp/terraform-plugin-framework/types"

type Webhook struct {
	ID         types.String `tfsdk:"id"`
	Secret     types.String `tfsdk:"secret"`
	URL        types.String `tfsdk:"url"`
	Active     types.Bool   `tfsdk:"active"`
	EventTypes types.List   `tfsdk:"event_types"`
}
