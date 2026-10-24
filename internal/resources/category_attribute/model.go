package category_attribute

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
)

type CategoryAttribute struct {
	CategoryId            types.String `tfsdk:"category_id"`
	AttributeDefinitionId types.String `tfsdk:"attribute_definition_id"`
	Value                 types.String `tfsdk:"value"`
	Mandatory             types.Bool   `tfsdk:"mandatory"`
	Propagate             types.Bool   `tfsdk:"propagate"`
	Locked                types.Bool   `tfsdk:"locked"`
}

func (c *CategoryAttribute) flags() cla.Flags {
	return cla.Flags{Mandatory: c.Mandatory, Propagate: c.Propagate, Locked: c.Locked}
}
