package attribute_definition

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

type AttributeDefinition struct {
	Id             types.String  `tfsdk:"id"`
	Name           types.String  `tfsdk:"name"`
	Number         types.String  `tfsdk:"number"`
	Description    types.String  `tfsdk:"description"`
	DataType       types.String  `tfsdk:"data_type"`
	ContentType    types.String  `tfsdk:"content_type"`
	CharacterSet   types.String  `tfsdk:"character_set"`
	ExternalSource types.Bool    `tfsdk:"external_source"`
	GroupID        types.String  `tfsdk:"group_id"`
	Internal       types.Bool    `tfsdk:"internal"`
	Unit           types.String  `tfsdk:"unit"`
	Restrictions   *Restrictions `tfsdk:"restrictions"`
}

type Restrictions struct {
	Enum  *EnumRestriction  `tfsdk:"enum"`
	Range *RangeRestriction `tfsdk:"range"`
	Text  *TextRestriction  `tfsdk:"text"`
}

func ToRestrictionsDto(r *Restrictions) *pim.RestrictionsDto {
	if r == nil {
		return nil
	}

	if r.Enum != nil {
		var values []pim.SelectAttributeValueDto
		for _, v := range *r.Enum.Values {
			values = append(values, pim.SelectAttributeValueDto{
				Metadata: v.Metadata.ValueStringPointer(),
				Number:   utils.KnownStringPointer(v.Number),
				Value:    v.Value.ValueStringPointer(),
				ValueId:  utils.KnownStringPointer(v.ValueId),
			})
		}

		return &pim.RestrictionsDto{
			Enum: &pim.SelectRestrictionsDto{
				Type:   utils.KnownStringPointer(r.Enum.Type),
				Values: &values,
			},
		}
	}

	if r.Range != nil {
		return &pim.RestrictionsDto{
			Range: &pim.RangeRestrictionsDto{
				Max:  r.Range.Max.ValueStringPointer(),
				Min:  r.Range.Min.ValueStringPointer(),
				Step: r.Range.Step.ValueStringPointer(),
			},
		}
	}

	if r.Text != nil {
		return &pim.RestrictionsDto{
			Text: &pim.TextRestrictionsDto{
				MaxLength:   r.Text.MaxLength.ValueInt32Pointer(),
				Pattern:     r.Text.Pattern.ValueStringPointer(),
				Whitespaces: r.Text.Whitespaces.ValueBoolPointer(),
			},
		}
	}

	return nil
}

type EnumRestriction struct {
	Type   types.String `tfsdk:"type"`
	Values *[]EnumValue `tfsdk:"values"`
}

type EnumValue struct {
	Metadata types.String `tfsdk:"metadata"`
	Number   types.String `tfsdk:"number"`
	Value    types.String `tfsdk:"value"`
	ValueId  types.String `tfsdk:"value_id"`
}

type RangeRestriction struct {
	Max  types.String `tfsdk:"max"`
	Min  types.String `tfsdk:"min"`
	Step types.String `tfsdk:"step"`
}

type TextRestriction struct {
	MaxLength   types.Int32  `tfsdk:"max_length"`
	Pattern     types.String `tfsdk:"pattern"`
	Whitespaces types.Bool   `tfsdk:"whitespaces"`
}
