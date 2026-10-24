package provider

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

const testCategoryId = "cat"

func newCategoryAttributeFake(t *testing.T) *fakebluestone.Server {
	fake := fakebluestone.New(t)
	fake.AddDefinition("d-text", map[string]any{"name": "Text", "dataType": "text"})
	fake.AddDefinition("d-unit", map[string]any{"name": "Unit", "dataType": "text"})
	fake.AddDefinition("d-column", map[string]any{"name": "Dimensions", "dataType": "column"})
	fake.AddDefinition("d-matrix", map[string]any{"name": "Size chart", "dataType": "matrix"})
	fake.AddDefinition("d-dictionary", map[string]any{"name": "Country", "dataType": "dictionary"})
	fake.AddDefinition("d-compound", map[string]any{
		"name": "Quantity", "isCompound": true,
		"subDefinitions": []any{map[string]any{"id": "d-unit"}},
	})
	return fake
}

func checkCategoryAttribute(fake *fakebluestone.Server, definitionId string, expected map[string]any) resource.TestCheckFunc {
	return func(*terraform.State) error {
		actual := fake.CategoryAttribute(testCategoryId, definitionId)
		if actual == nil {
			return fmt.Errorf("%s is not assigned", definitionId)
		}
		for key, want := range expected {
			if got := actual[key]; !reflect.DeepEqual(got, want) {
				return fmt.Errorf("%s: expected %s to be %#v, got %#v", definitionId, key, want, got)
			}
		}
		return nil
	}
}

func TestCategoryAttributes_AllTypes(t *testing.T) {
	fake := newCategoryAttributeFake(t)

	runTest(t,
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_category_attribute" "text" {
  category_id             = "cat"
  attribute_definition_id = "d-text"
  value                   = "hello"
  mandatory               = true
  locked                  = true
}

resource "bluestonepim_category_column_attribute" "column" {
  category_id             = "cat"
  attribute_definition_id = "d-column"
  values                  = { c1 = "10", c2 = "20" }
  propagate               = true
}

resource "bluestonepim_category_matrix_attribute" "matrix" {
  category_id             = "cat"
  attribute_definition_id = "d-matrix"
  values                  = { c1 = { r1 = "92" } }
}

resource "bluestonepim_category_dictionary_attribute" "dictionary" {
  category_id             = "cat"
  attribute_definition_id = "d-dictionary"
  value_ids               = ["v1", "v2"]
  mandatory               = true
}

resource "bluestonepim_category_compound_attribute" "compound" {
  category_id             = "cat"
  attribute_definition_id = "d-compound"
  propagate               = true
}

# Assigned together with the compound attribute
resource "bluestonepim_category_attribute" "unit" {
  category_id             = "cat"
  attribute_definition_id = "d-unit"
  value                   = "kg"

  depends_on = [bluestonepim_category_compound_attribute.compound]
}
`),
			Check: resource.ComposeTestCheckFunc(
				checkCategoryAttribute(fake, "d-text", map[string]any{"value": "hello", "mandatory": true, "locked": true}),
				checkCategoryAttribute(fake, "d-column", map[string]any{"column": map[string]any{"c1": "10", "c2": "20"}, "propagate": true}),
				checkCategoryAttribute(fake, "d-matrix", map[string]any{"matrix": map[string]any{"c1": map[string]any{"r1": "92"}}}),
				checkCategoryAttribute(fake, "d-dictionary", map[string]any{"dictionary": []any{"v1", "v2"}, "mandatory": true}),
				checkCategoryAttribute(fake, "d-compound", map[string]any{"propagate": true}),
				checkCategoryAttribute(fake, "d-unit", map[string]any{"value": "kg"}),
			),
		},
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_category_attribute" "text" {
  category_id             = "cat"
  attribute_definition_id = "d-text"
}

resource "bluestonepim_category_column_attribute" "column" {
  category_id             = "cat"
  attribute_definition_id = "d-column"
  values                  = { c1 = "10", c2 = "30" }
}

resource "bluestonepim_category_matrix_attribute" "matrix" {
  category_id             = "cat"
  attribute_definition_id = "d-matrix"
  values                  = { c1 = { r1 = "92" }, c2 = { r1 = "100" } }
  mandatory               = true
}

resource "bluestonepim_category_dictionary_attribute" "dictionary" {
  category_id             = "cat"
  attribute_definition_id = "d-dictionary"
}

resource "bluestonepim_category_compound_attribute" "compound" {
  category_id             = "cat"
  attribute_definition_id = "d-compound"
}

resource "bluestonepim_category_attribute" "unit" {
  category_id             = "cat"
  attribute_definition_id = "d-unit"
  value                   = "g"

  depends_on = [bluestonepim_category_compound_attribute.compound]
}
`),
			Check: resource.ComposeTestCheckFunc(
				checkCategoryAttribute(fake, "d-text", map[string]any{"value": nil, "mandatory": false, "locked": false}),
				checkCategoryAttribute(fake, "d-column", map[string]any{"column": map[string]any{"c1": "10", "c2": "30"}, "propagate": false}),
				checkCategoryAttribute(fake, "d-matrix", map[string]any{"mandatory": true}),
				checkCategoryAttribute(fake, "d-dictionary", map[string]any{"dictionary": []any{}, "mandatory": false}),
				checkCategoryAttribute(fake, "d-compound", map[string]any{"propagate": false}),
				checkCategoryAttribute(fake, "d-unit", map[string]any{"value": "g"}),
			),
		},
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_category_column_attribute" "column" {
  category_id             = "cat"
  attribute_definition_id = "d-column"
  values                  = { c1 = "10", c2 = "30" }
}
`),
			ResourceName:                         "bluestonepim_category_column_attribute.column",
			ImportState:                          true,
			ImportStateId:                        "cat/d-column",
			ImportStateVerify:                    true,
			ImportStateVerifyIdentifierAttribute: "attribute_definition_id",
		},
	)
}

func TestCategoryAttribute_LockedRequiresMandatory(t *testing.T) {
	fake := newCategoryAttributeFake(t)

	runTest(t, resource.TestStep{
		Config: config(fake, `
resource "bluestonepim_category_attribute" "text" {
  category_id             = "cat"
  attribute_definition_id = "d-text"
  locked                  = true
}
`),
		ExpectError: regexpMustCompile(`A locked attribute is always mandatory`),
	})
}
