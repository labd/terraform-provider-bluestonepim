package provider

import (
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

func TestAttributeGroup_NumberAndOrderByNameUpdatedInPlace(t *testing.T) {
	fake := fakebluestone.New(t)

	runTest(t,
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_attribute_group" "general" {
  name   = "General"
  number = "general"
}
`),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("bluestonepim_attribute_group.general", "number", "general"),
				// Not configured, so the value from Bluestone
				resource.TestCheckResourceAttr("bluestonepim_attribute_group.general", "order_by_name", "true"),
			),
		},
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_attribute_group" "general" {
  name          = "General"
  number        = "general-info"
  order_by_name = false
}
`),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("bluestonepim_attribute_group.general", plancheck.ResourceActionUpdate),
				},
			},
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("bluestonepim_attribute_group.general", "number", "general-info"),
				resource.TestCheckResourceAttr("bluestonepim_attribute_group.general", "order_by_name", "false"),
			),
		},
	)
}

func TestAttributeGroupOrder(t *testing.T) {
	fake := fakebluestone.New(t)
	// A group not managed by Terraform, which ends up after the ordered groups
	fake.AddAttributeGroup("g-unmanaged", "Unmanaged")

	groups := `
resource "bluestonepim_attribute_group" "general" {
  name = "General"
}

resource "bluestonepim_attribute_group" "technical" {
  name = "Technical"
}
`
	expectOrder := func(addresses ...string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			var expected []string
			for _, address := range addresses {
				if address == "g-unmanaged" {
					expected = append(expected, address)
					continue
				}
				id, err := stateValue(s, address, "id")
				if err != nil {
					return err
				}
				expected = append(expected, id)
			}
			if actual := fake.AttributeGroupOrder(); !slices.Equal(actual, expected) {
				return fmt.Errorf("expected group order %v, got %v", expected, actual)
			}
			return nil
		}
	}

	runTest(t,
		resource.TestStep{
			Config: config(fake, groups+`
resource "bluestonepim_attribute_group_order" "this" {
  group_ids = [
    bluestonepim_attribute_group.technical.id,
    bluestonepim_attribute_group.general.id,
  ]
}
`),
			Check: expectOrder("bluestonepim_attribute_group.technical", "bluestonepim_attribute_group.general", "g-unmanaged"),
		},
		resource.TestStep{
			Config: config(fake, groups+`
resource "bluestonepim_attribute_group_order" "this" {
  group_ids = [
    bluestonepim_attribute_group.general.id,
    bluestonepim_attribute_group.technical.id,
  ]
}
`),
			Check: expectOrder("bluestonepim_attribute_group.general", "bluestonepim_attribute_group.technical", "g-unmanaged"),
		},
	)
}

func TestAttributeGroupDefinitionOrder(t *testing.T) {
	fake := fakebluestone.New(t)

	resources := func(orderByName bool, order string) string {
		return config(fake, fmt.Sprintf(`
resource "bluestonepim_attribute_group" "general" {
  name          = "General"
  order_by_name = %t
}

resource "bluestonepim_attribute_definition" "brand" {
  name      = "Brand"
  data_type = "text"
  group_id  = bluestonepim_attribute_group.general.id
}

resource "bluestonepim_attribute_definition" "ean" {
  name      = "EAN"
  data_type = "text"
  group_id  = bluestonepim_attribute_group.general.id
}

resource "bluestonepim_attribute_definition" "series" {
  name      = "Series"
  data_type = "text"
  group_id  = bluestonepim_attribute_group.general.id
}

resource "bluestonepim_attribute_group_definition_order" "general" {
  group_id       = bluestonepim_attribute_group.general.id
  definition_ids = [%s]
}
`, orderByName, order))
	}
	expectOrder := func(names ...string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			groupId, err := stateValue(s, "bluestonepim_attribute_group.general", "id")
			if err != nil {
				return err
			}
			var expected []string
			for _, name := range names {
				id, err := stateValue(s, "bluestonepim_attribute_definition."+name, "id")
				if err != nil {
					return err
				}
				expected = append(expected, id)
			}
			if actual := fake.GroupDefinitionOrder(groupId); !slices.Equal(actual, expected) {
				return fmt.Errorf("expected definition order %v, got %v", expected, actual)
			}
			return nil
		}
	}

	runTest(t,
		// Ordered by name, a manual order is rejected
		resource.TestStep{
			Config: resources(true, `
    bluestonepim_attribute_definition.series.id,
    bluestonepim_attribute_definition.brand.id,`),
			ExpectError: regexpMustCompile("Attribute group is ordered by name"),
		},
		resource.TestStep{
			Config: resources(false, `
    bluestonepim_attribute_definition.series.id,
    bluestonepim_attribute_definition.brand.id,`),
			Check: expectOrder("series", "brand", "ean"),
		},
		resource.TestStep{
			Config: resources(false, `
    bluestonepim_attribute_definition.ean.id,
    bluestonepim_attribute_definition.series.id,
    bluestonepim_attribute_definition.brand.id,`),
			Check: expectOrder("ean", "series", "brand"),
		},
	)
}
