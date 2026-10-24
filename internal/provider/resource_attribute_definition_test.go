package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

func TestAttributeDefinition_SelectValueIdsAreKept(t *testing.T) {
	fake := fakebluestone.New(t)
	ids := map[string]string{}

	runTest(t,
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_attribute_definition" "colour" {
  name      = "Colour"
  data_type = "single_select"
  restrictions = {
    enum = {
      values = [
        { value = "Red", number = "red" },
        { value = "Blue", number = "blue" },
      ]
    }
  }
}
`),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttrSet("bluestonepim_attribute_definition.colour", "restrictions.enum.values.0.value_id"),
				saveValueIds(ids, "Red", "Blue"),
			),
		},
		// Reorder the values, insert a new one in between and rename one: the
		// existing values keep their ID, matched by number.
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_attribute_definition" "colour" {
  name      = "Colour"
  data_type = "single_select"
  restrictions = {
    enum = {
      values = [
        { value = "Blue", number = "blue" },
        { value = "Green", number = "green" },
        { value = "Rood", number = "red" },
      ]
    }
  }
}
`),
			Check: resource.ComposeTestCheckFunc(
				checkValueId(ids, 0, "Blue"),
				checkValueId(ids, 2, "Red"), // renamed to Rood
			),
		},
	)
}

// Labels don't have to be unique: two values with the same label keep their
// own ID when reordered.
func TestAttributeDefinition_DuplicateLabels(t *testing.T) {
	fake := fakebluestone.New(t)
	fake.AddDefinition("def-wheel", map[string]any{
		"name": "Option", "number": "option", "dataType": "single_select",
		"externalSource": false, "internal": false,
		"restrictions": map[string]any{"enum": map[string]any{"values": []any{
			map[string]any{"value": `20" Wheel`, "valueId": "id-1", "number": "wheel-a"},
			map[string]any{"value": `20" Wheel`, "valueId": "id-2", "number": "wheel-b"},
		}}},
	})

	resourceConfig := func(first, second string) string {
		return config(fake, fmt.Sprintf(`
resource "bluestonepim_attribute_definition" "option" {
  name      = "Option"
  number    = "option"
  data_type = "single_select"
  restrictions = {
    enum = {
      values = [
        { value = "20\" Wheel", number = "%s" },
        { value = "20\" Wheel", number = "%s" },
      ]
    }
  }
}
`, first, second))
	}

	runTest(t,
		resource.TestStep{
			Config: resourceConfig("wheel-a", "wheel-b") + `
import {
  to = bluestonepim_attribute_definition.option
  id = "def-wheel"
}
`,
		},
		resource.TestStep{
			Config: resourceConfig("wheel-b", "wheel-a"),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("bluestonepim_attribute_definition.option", "restrictions.enum.values.0.value_id", "id-2"),
				resource.TestCheckResourceAttr("bluestonepim_attribute_definition.option", "restrictions.enum.values.1.value_id", "id-1"),
			),
		},
	)
}

func saveValueIds(ids map[string]string, values ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i, value := range values {
			id, err := stateValue(s, "bluestonepim_attribute_definition.colour", fmt.Sprintf("restrictions.enum.values.%d.value_id", i))
			if err != nil {
				return err
			}
			ids[value] = id
		}
		return nil
	}
}

func checkValueId(ids map[string]string, index int, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id, err := stateValue(s, "bluestonepim_attribute_definition.colour", fmt.Sprintf("restrictions.enum.values.%d.value_id", index))
		if err != nil {
			return err
		}
		if id != ids[value] {
			return fmt.Errorf("expected %s to keep value_id %s, got %s", value, ids[value], id)
		}
		return nil
	}
}

// Definitions without a content type (all data types except formatted text)
// should import without changes.
func TestAttributeDefinition_ImportWithoutChanges(t *testing.T) {
	fake := fakebluestone.New(t)
	fake.AddDefinition("def-text", map[string]any{
		"name": "Bars", "number": "bars", "dataType": "text",
		"externalSource": false, "internal": false,
	})
	fake.AddDefinition("def-select", map[string]any{
		"name": "Size", "number": "size", "dataType": "single_select",
		"externalSource": false, "internal": false,
		"restrictions": map[string]any{"enum": map[string]any{"values": []any{
			map[string]any{"value": "S", "valueId": "id-s", "number": "s"},
			map[string]any{"value": "M", "valueId": "id-m", "number": "m"},
		}}},
	})

	resourceConfig := config(fake, `
resource "bluestonepim_attribute_definition" "bars" {
  name      = "Bars"
  number    = "bars"
  data_type = "text"
}

resource "bluestonepim_attribute_definition" "size" {
  name      = "Size"
  number    = "size"
  data_type = "single_select"
  restrictions = {
    enum = {
      values = [
        { value = "S" },
        { value = "M" },
      ]
    }
  }
}
`)

	runTest(t,
		resource.TestStep{
			Config: resourceConfig + `
import {
  to = bluestonepim_attribute_definition.bars
  id = "def-text"
}

import {
  to = bluestonepim_attribute_definition.size
  id = "def-select"
}
`,
		},
		resource.TestStep{
			Config:   resourceConfig,
			PlanOnly: true,
		},
	)
}

func TestAttributeDefinition_DeletedOutsideTerraform(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	resourceConfig := config(fake, `
resource "bluestonepim_attribute_definition" "bars" {
  name      = "Bars"
  data_type = "text"
}
`)

	runTest(t,
		resource.TestStep{
			Config: resourceConfig,
			Check: func(s *terraform.State) (err error) {
				id, err = stateValue(s, "bluestonepim_attribute_definition.bars", "id")
				return err
			},
		},
		resource.TestStep{
			PreConfig:          func() { fake.RemoveDefinition(id) },
			Config:             resourceConfig,
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		},
	)
}

func TestAttributeDefinition_RestrictionsMustMatchDataType(t *testing.T) {
	fake := fakebluestone.New(t)

	cases := map[string]string{
		"pattern on text": `
  data_type = "text"
  restrictions = { text = { pattern = "^[0-9]+$" } }`,
		"range on text": `
  data_type = "text"
  restrictions = { range = { min = "1" } }`,
		"enum on text": `
  data_type = "text"
  restrictions = { enum = { values = [{ value = "A" }] } }`,
		"content type on text": `
  data_type    = "text"
  content_type = "text/html"`,
		"select without values": `
  data_type = "single_select"`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			runTest(t, resource.TestStep{
				Config: config(fake, `
resource "bluestonepim_attribute_definition" "invalid" {
  name = "Invalid"
`+body+`
}
`),
				ExpectError: regexpMustCompile(`Restriction not supported for data type|Missing select values`),
			})
		})
	}

	// Valid combinations
	runTest(t, resource.TestStep{
		Config: config(fake, `
resource "bluestonepim_attribute_definition" "gtin" {
  name         = "GTIN"
  data_type    = "pattern"
  restrictions = { text = { pattern = "^[0-9]{13}$" } }
}

resource "bluestonepim_attribute_definition" "title" {
  name         = "Title"
  data_type    = "text"
  restrictions = { text = { max_length = 100 } }
}

resource "bluestonepim_attribute_definition" "weight" {
  name         = "Weight"
  data_type    = "decimal"
  restrictions = { range = { min = "0" } }
}

resource "bluestonepim_attribute_definition" "body" {
  name         = "Body"
  data_type    = "formatted_text"
  content_type = "text/html"
}
`),
	})
}

func TestAttributeDefinition_DataTypeCannotBeChanged(t *testing.T) {
	fake := fakebluestone.New(t)

	runTest(t,
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_attribute_definition" "brand" {
  name      = "Brand"
  data_type = "text"
}
`),
		},
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_attribute_definition" "brand" {
  name      = "Brand"
  data_type = "multiline"
}
`),
			ExpectError: regexpMustCompile("Data type cannot be changed"),
		},
	)
}
