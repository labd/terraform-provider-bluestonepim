package matrix_attribute_definition

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/labd/terraform-provider-bluestonepim/internal/resources/attribute_definition"
)

type MatrixAttributeDefinition struct {
	Id             types.String                             `tfsdk:"id"`
	Name           types.String                             `tfsdk:"name"`
	Number         types.String                             `tfsdk:"number"`
	Description    types.String                             `tfsdk:"description"`
	ExternalSource types.Bool                               `tfsdk:"external_source"`
	GroupID        types.String                             `tfsdk:"group_id"`
	Internal       types.Bool                               `tfsdk:"internal"`
	Columns        []attribute_definition.IdentifiableValue `tfsdk:"columns"`
	Rows           []attribute_definition.IdentifiableValue `tfsdk:"rows"`
}
