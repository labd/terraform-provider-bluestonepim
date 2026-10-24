package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

func TestCategory_Lifecycle(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	runTest(t,
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_category" "bikes" {
  name        = "Bikes"
  description = "All our bikes"
}
`),
			Check: resource.ComposeTestCheckFunc(
				// The number defaults to the ID
				resource.TestCheckResourceAttrPair("bluestonepim_category.bikes", "number", "bluestonepim_category.bikes", "id"),
				func(s *terraform.State) (err error) {
					id, err = stateValue(s, "bluestonepim_category.bikes", "id")
					return err
				},
			),
		},
		// Removing the description clears it, the API then returns an empty
		// description which should not cause a diff.
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_category" "bikes" {
  name = "Bikes"
}
`),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckNoResourceAttr("bluestonepim_category.bikes", "description"),
				func(*terraform.State) error {
					if v := fake.CategoryValues(id, "en"); v["description"] != "" {
						return fmt.Errorf("expected description to be cleared, got %q", v["description"])
					}
					return nil
				},
			),
		},
		resource.TestStep{
			PreConfig: func() { fake.DeleteCategory(id) },
			Config: config(fake, `
resource "bluestonepim_category" "bikes" {
  name = "Bikes"
}
`),
			PlanOnly:           true,
			ExpectNonEmptyPlan: true,
		},
	)
}

func TestCategoryContext_Lifecycle(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	category := `
resource "bluestonepim_category" "bikes" {
  name        = "Bikes"
  description = "All our bikes"
}
`
	runTest(t,
		resource.TestStep{
			Config: config(fake, category+`
resource "bluestonepim_category_context" "bikes_nl" {
  category_id = bluestonepim_category.bikes.id
  context_id  = "L1"
  name        = "Fietsen"
  description = "Al onze fietsen"
}
`),
			Check: func(s *terraform.State) (err error) {
				id, err = stateValue(s, "bluestonepim_category.bikes", "id")
				if err != nil {
					return err
				}
				if v := fake.CategoryValues(id, "L1"); v["name"] != "Fietsen" {
					return fmt.Errorf("expected the name to be set for L1, got %v", v)
				}
				if v := fake.CategoryValues(id, "en"); v["name"] != "Bikes" {
					return fmt.Errorf("expected the default context to be untouched, got %v", v)
				}
				return nil
			},
		},
		// Removing the description clears it for the context only, so it falls
		// back to the default context.
		resource.TestStep{
			Config: config(fake, category+`
resource "bluestonepim_category_context" "bikes_nl" {
  category_id = bluestonepim_category.bikes.id
  context_id  = "L1"
  name        = "Fietsen"
}
`),
			Check: func(*terraform.State) error {
				if v := fake.CategoryValues(id, "L1"); v["description"] != "" {
					return fmt.Errorf("expected the L1 description to be cleared, got %v", v)
				}
				if v := fake.CategoryValues(id, "en"); v["description"] != "All our bikes" {
					return fmt.Errorf("expected the default description to be untouched, got %v", v)
				}
				return nil
			},
		},
		// Removing the resource clears the context values
		resource.TestStep{
			Config: config(fake, category),
			Check: func(*terraform.State) error {
				if v := fake.CategoryValues(id, "L1"); len(v) != 0 {
					return fmt.Errorf("expected the L1 values to be cleared, got %v", v)
				}
				return nil
			},
		},
		resource.TestStep{
			Config:        config(fake, category),
			ResourceName:  "bluestonepim_category.bikes",
			ImportState:   true,
			ImportStateId: "",
			ImportStateIdFunc: func(*terraform.State) (string, error) {
				return id, nil
			},
			ImportStateVerify: true,
		},
	)
}
