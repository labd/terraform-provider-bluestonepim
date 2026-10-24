package compound_attribute_definition

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

type CompoundAttributeDefinition struct {
	Id             types.String    `tfsdk:"id"`
	Name           types.String    `tfsdk:"name"`
	Number         types.String    `tfsdk:"number"`
	Description    types.String    `tfsdk:"description"`
	ExternalSource types.Bool      `tfsdk:"external_source"`
	GroupID        types.String    `tfsdk:"group_id"`
	Internal       types.Bool      `tfsdk:"internal"`
	Format         []types.String  `tfsdk:"format"`
	SubDefinitions []SubDefinition `tfsdk:"sub_definitions"`
}

type SubDefinition struct {
	Id       types.String `tfsdk:"id"`
	Compound types.Bool   `tfsdk:"compound"`
}

func (c *CompoundAttributeDefinition) toDto() pim.CompoundAttributeDefinitionRequest {
	subDefinitions := make([]pim.CompoundAttributeSubDefinitionDto, 0, len(c.SubDefinitions))
	for _, s := range c.SubDefinitions {
		subDefinitions = append(subDefinitions, pim.CompoundAttributeSubDefinitionDto{
			Id:       s.Id.ValueStringPointer(),
			Compound: s.Compound.ValueBoolPointer(),
		})
	}

	var format *[]string
	if c.Format != nil {
		values := make([]string, 0, len(c.Format))
		for _, f := range c.Format {
			values = append(values, f.ValueString())
		}
		format = &values
	}

	return pim.CompoundAttributeDefinitionRequest{
		ExternalSource: c.ExternalSource.ValueBoolPointer(),
		Format:         format,
		GroupId:        c.GroupID.ValueStringPointer(),
		Internal:       c.Internal.ValueBoolPointer(),
		Name:           c.Name.ValueStringPointer(),
		Number:         utils.KnownStringPointer(c.Number),
		SubDefinitions: &subDefinitions,
	}
}

func fromResponse(r *pim.CompoundAttributeDefinitionResponse) *CompoundAttributeDefinition {
	result := &CompoundAttributeDefinition{
		Id:             types.StringPointerValue(r.Id),
		Name:           types.StringValue(r.Name),
		Number:         types.StringPointerValue(r.Number),
		Description:    types.StringPointerValue(r.Description),
		ExternalSource: types.BoolPointerValue(r.ExternalSource),
		GroupID:        types.StringPointerValue(r.GroupId),
		Internal:       types.BoolPointerValue(r.Internal),
	}

	if r.Format != nil && len(*r.Format) > 0 {
		result.Format = make([]types.String, 0, len(*r.Format))
		for _, f := range *r.Format {
			result.Format = append(result.Format, types.StringValue(f))
		}
	}

	if r.SubDefinitions != nil {
		result.SubDefinitions = make([]SubDefinition, 0, len(*r.SubDefinitions))
		for _, s := range *r.SubDefinitions {
			compound := false
			if s.Compound != nil {
				compound = *s.Compound
			}
			result.SubDefinitions = append(result.SubDefinitions, SubDefinition{
				Id:       types.StringPointerValue(s.Id),
				Compound: types.BoolValue(compound),
			})
		}
	}

	return result
}
