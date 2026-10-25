package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

func TestAttributeDefinitionContext_Lifecycle(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	definition := `
resource "bluestonepim_attribute_definition" "ean" {
  name        = "EAN"
  description = "European Article Number"
  data_type   = "text"
}
`
	runTest(t,
		resource.TestStep{
			Config: config(fake, definition+`
resource "bluestonepim_attribute_definition_context" "ean_nl" {
  definition_id = bluestonepim_attribute_definition.ean.id
  context_id    = "l3730"
  name          = "EAN-code"
  description   = "Europees artikelnummer"
}
`),
			Check: func(s *terraform.State) (err error) {
				id, err = stateValue(s, "bluestonepim_attribute_definition.ean", "id")
				if err != nil {
					return err
				}
				if v := fake.DefinitionValues(id, "l3730"); v["name"] != "EAN-code" || v["description"] != "Europees artikelnummer" {
					return fmt.Errorf("expected the values to be set for l3730, got %v", v)
				}
				if d := fake.Definition(id); d["name"] != "EAN" {
					return fmt.Errorf("expected the default name to be untouched, got %v", d["name"])
				}
				return nil
			},
		},
		// Removing the description clears it for the context only
		resource.TestStep{
			Config: config(fake, definition+`
resource "bluestonepim_attribute_definition_context" "ean_nl" {
  definition_id = bluestonepim_attribute_definition.ean.id
  context_id    = "l3730"
  name          = "EAN-code"
}
`),
			Check: func(*terraform.State) error {
				if v := fake.DefinitionValues(id, "l3730"); v["description"] != nil {
					return fmt.Errorf("expected the l3730 description to be cleared, got %v", v)
				}
				if d := fake.Definition(id); d["description"] != "European Article Number" {
					return fmt.Errorf("expected the default description to be untouched, got %v", d["description"])
				}
				return nil
			},
		},
		// Removing the resource clears the context values
		resource.TestStep{
			Config: config(fake, definition),
			Check: func(*terraform.State) error {
				if v := fake.DefinitionValues(id, "l3730"); len(v) != 0 {
					return fmt.Errorf("expected the l3730 values to be cleared, got %v", v)
				}
				return nil
			},
		},
	)
}

func TestAttributeDefinitionContext_ImportWithoutChanges(t *testing.T) {
	fake := fakebluestone.New(t)
	fake.AddDefinition("def-ean", map[string]any{
		"name": "EAN", "dataType": "text",
		"contextValues": map[string]map[string]any{"l3730": {"name": "EAN-code"}},
	})

	resourceConfig := config(fake, `
resource "bluestonepim_attribute_definition_context" "ean_nl" {
  definition_id = "def-ean"
  context_id    = "l3730"
  name          = "EAN-code"
}
`)
	runTest(t,
		resource.TestStep{
			Config: resourceConfig + `
import {
  to = bluestonepim_attribute_definition_context.ean_nl
  id = "def-ean/l3730"
}
`,
		},
		resource.TestStep{Config: resourceConfig, PlanOnly: true},
	)
}

func TestAttributeGroupContext_Lifecycle(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	group := `
resource "bluestonepim_attribute_group" "general" {
  name = "General"
}
`
	runTest(t,
		resource.TestStep{
			Config: config(fake, group+`
resource "bluestonepim_attribute_group_context" "general_nl" {
  group_id   = bluestonepim_attribute_group.general.id
  context_id = "l3730"
  name       = "Algemeen"
}
`),
			Check: func(s *terraform.State) (err error) {
				id, err = stateValue(s, "bluestonepim_attribute_group.general", "id")
				if err != nil {
					return err
				}
				if n := fake.AttributeGroupName(id, "l3730"); n != "Algemeen" {
					return fmt.Errorf("expected the l3730 name to be set, got %q", n)
				}
				if n := fake.AttributeGroupName(id, "en"); n != "General" {
					return fmt.Errorf("expected the default name to be untouched, got %q", n)
				}
				return nil
			},
		},
		resource.TestStep{
			Config: config(fake, group+`
resource "bluestonepim_attribute_group_context" "general_nl" {
  group_id   = bluestonepim_attribute_group.general.id
  context_id = "l3730"
  name       = "Algemene gegevens"
}
`),
			Check: func(*terraform.State) error {
				if n := fake.AttributeGroupName(id, "l3730"); n != "Algemene gegevens" {
					return fmt.Errorf("expected the l3730 name to be updated, got %q", n)
				}
				return nil
			},
		},
		resource.TestStep{
			ResourceName:                         "bluestonepim_attribute_group_context.general_nl",
			ImportState:                          true,
			ImportStateVerify:                    true,
			ImportStateIdFunc:                    func(*terraform.State) (string, error) { return id + "/l3730", nil },
			ImportStateVerifyIdentifierAttribute: "group_id",
			Config: config(fake, group+`
resource "bluestonepim_attribute_group_context" "general_nl" {
  group_id   = bluestonepim_attribute_group.general.id
  context_id = "l3730"
  name       = "Algemene gegevens"
}
`),
		},
	)
}
