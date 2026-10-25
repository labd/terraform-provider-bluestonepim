package dictionary_attribute_definition

import "github.com/hashicorp/terraform-plugin-framework/types"

type DictionaryAttributeDefinition struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Number             types.String `tfsdk:"number"`
	Description        types.String `tfsdk:"description"`
	ExternalSource     types.Bool   `tfsdk:"external_source"`
	GroupID            types.String `tfsdk:"group_id"`
	Internal           types.Bool   `tfsdk:"internal"`
	SelectedValueLimit types.Int32  `tfsdk:"selected_values_limit"`
	ParentDefinitionId types.String `tfsdk:"parent_definition_id"`
	ContextAware       types.Bool   `tfsdk:"context_aware"`
}
