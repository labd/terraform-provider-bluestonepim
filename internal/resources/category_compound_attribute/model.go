package category_compound_attribute

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
)

type CategoryCompoundAttribute struct {
	CategoryId            types.String `tfsdk:"category_id"`
	AttributeDefinitionId types.String `tfsdk:"attribute_definition_id"`
	Mandatory             types.Bool   `tfsdk:"mandatory"`
	Propagate             types.Bool   `tfsdk:"propagate"`
}

func (c *CategoryCompoundAttribute) flags() cla.Flags {
	return cla.Flags{Mandatory: c.Mandatory, Propagate: c.Propagate}
}
