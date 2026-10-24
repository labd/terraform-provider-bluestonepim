package category_column_attribute

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
)

type CategoryColumnAttribute struct {
	CategoryId            types.String `tfsdk:"category_id"`
	AttributeDefinitionId types.String `tfsdk:"attribute_definition_id"`
	Values                types.Map    `tfsdk:"values"`
	Mandatory             types.Bool   `tfsdk:"mandatory"`
	Propagate             types.Bool   `tfsdk:"propagate"`
}

func (c *CategoryColumnAttribute) flags() cla.Flags {
	return cla.Flags{Mandatory: c.Mandatory, Propagate: c.Propagate}
}

var valuesType = types.StringType
