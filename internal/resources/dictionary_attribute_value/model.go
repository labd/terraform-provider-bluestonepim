package dictionary_attribute_value

import "github.com/hashicorp/terraform-plugin-framework/types"

type DictionaryAttributeValue struct {
	Id                    types.String `tfsdk:"id"`
	AttributeDefinitionId types.String `tfsdk:"attribute_definition_id"`
	Value                 types.String `tfsdk:"value"`
	Number                types.String `tfsdk:"number"`
}
