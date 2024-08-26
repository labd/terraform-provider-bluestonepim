package attribute_definition

import "github.com/hashicorp/terraform-plugin-framework/types"

type AttributeDefinition struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Number         types.String `tfsdk:"number"`
	DataType       types.String `tfsdk:"data_type"`
	ContentType    types.String `tfsdk:"content_type"`
	CharacterSet   types.String `tfsdk:"character_set"`
	ExternalSource types.Bool   `tfsdk:"external_source"`
	GroupID        types.String `tfsdk:"group_id"`
	Internal       types.Bool   `tfsdk:"internal"`
	Unit           types.String `tfsdk:"unit"`
	Restrictions   Restrictions `tfsdk:"restrictions"`
}

type Restrictions struct {
	types.Object

	Column ColumnRestriction `tfsdk:"column"`
	Enum   EnumRestriction   `tfsdk:"enum"`
	Matrix types.Object      `tfsdk:"matrix"`
	Range  types.Object      `tfsdk:"range"`
	Text   types.Object      `tfsdk:"text"`
}

type ColumnRestriction struct {
	types.Object

	Columns types.List `tfsdk:"columns"`
}

type Column struct {
	types.Object

	ID    types.String `tfsdk:"id"`
	Value types.String `tfsdk:"value"`
}

type EnumRestriction struct {
	types.Object

	Type   types.String `tfsdk:"type"`
	Values types.List   `tfsdk:"values"`
}

type EnumValue struct {
	types.Object

	Metadata types.String `tfsdk:"metadata"`
	Number   types.String `tfsdk:"number"`
	Value    types.String `tfsdk:"value"`
	ValueId  types.String `tfsdk:"valueId"`
}
