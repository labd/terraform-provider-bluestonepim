package cla

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// SchemaAttributes returns the attributes shared by all CLA resources. The
// lock flag is not supported for column, matrix and compound definitions.
func SchemaAttributes(supportsLocked bool) map[string]schema.Attribute {
	attributes := map[string]schema.Attribute{
		"category_id": schema.StringAttribute{
			MarkdownDescription: "The ID of the category.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"attribute_definition_id": schema.StringAttribute{
			MarkdownDescription: "The ID of the attribute definition.",
			Required:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"mandatory": schema.BoolAttribute{
			MarkdownDescription: "Require all products in this category to have a value for this attribute.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
		"propagate": schema.BoolAttribute{
			MarkdownDescription: "Propagate the attribute and its value to all subcategories and their products.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		},
	}

	if supportsLocked {
		attributes["mandatory"] = schema.BoolAttribute{
			MarkdownDescription: "Require all products in this category to have a value for this attribute. " +
				"Must be `true` when `locked` is set, since locking also makes the attribute mandatory.",
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		}
		attributes["locked"] = schema.BoolAttribute{
			MarkdownDescription: "Require the products in this category to keep the value of this attribute.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
		}
	}

	return attributes
}
