package category_dictionary_attribute

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/cla"
)

type CategoryDictionaryAttribute struct {
	CategoryId            types.String `tfsdk:"category_id"`
	AttributeDefinitionId types.String `tfsdk:"attribute_definition_id"`
	ValueIds              types.Set    `tfsdk:"value_ids"`
	Mandatory             types.Bool   `tfsdk:"mandatory"`
	Propagate             types.Bool   `tfsdk:"propagate"`
	Locked                types.Bool   `tfsdk:"locked"`
}

func (c *CategoryDictionaryAttribute) flags() cla.Flags {
	return cla.Flags{Mandatory: c.Mandatory, Propagate: c.Propagate, Locked: c.Locked}
}
