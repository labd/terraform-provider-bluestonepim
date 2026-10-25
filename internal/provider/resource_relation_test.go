package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

func TestRelation_Lifecycle(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	runTest(t,
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_relation" "successor" {
  name         = "Superseded by"
  reverse_name = "Successor of"
  number       = "successor"
  description  = "The product that replaces this one."
}
`),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("bluestonepim_relation.successor", "direction", "ONE_WAY"),
				resource.TestCheckResourceAttr("bluestonepim_relation.successor", "quantity_enabled", "false"),
				resource.TestCheckResourceAttr("bluestonepim_relation.successor", "force_destroy", "false"),
				func(s *terraform.State) (err error) {
					id, err = stateValue(s, "bluestonepim_relation.successor", "id")
					return err
				},
			),
		},
		// Rename, remove the description and switch the direction, in place
		resource.TestStep{
			Config: config(fake, `
resource "bluestonepim_relation" "successor" {
  name         = "Replaced by"
  reverse_name = "Replaces"
  number       = "successor"
  direction    = "TWO_WAY"
}
`),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("bluestonepim_relation.successor", plancheck.ResourceActionUpdate),
				},
			},
			Check: func(*terraform.State) error {
				r := fake.Relation(id)
				if r["name"] != "Replaced by" || r["direction"] != "TWO_WAY" || r["description"] != nil {
					return fmt.Errorf("unexpected relation %v", r)
				}
				// Updates use the partial update, not the deprecated PUT
				for _, req := range fake.Requests() {
					if req.Method == "PUT" {
						return fmt.Errorf("unexpected PUT %s", req.Path)
					}
				}
				return nil
			},
		},
	)
}

func TestRelation_ImportWithoutChanges(t *testing.T) {
	fake := fakebluestone.New(t)
	fake.AddRelation("696a371e958d1c0304e3d3ee", map[string]any{
		"name": "is a variant of", "number": "variant", "direction": "ONE_WAY", "quantityEnabled": false,
	})

	resourceConfig := config(fake, `
resource "bluestonepim_relation" "variant" {
  name   = "is a variant of"
  number = "variant"
}
`)
	runTest(t,
		resource.TestStep{
			Config: resourceConfig + `
import {
  to = bluestonepim_relation.variant
  id = "696a371e958d1c0304e3d3ee"
}
`,
		},
		resource.TestStep{
			Config:   resourceConfig,
			PlanOnly: true,
		},
	)
}

func TestRelation_DeleteInUseRequiresForceDestroy(t *testing.T) {
	fake := fakebluestone.New(t)
	var id string

	relation := func(forceDestroy bool) string {
		return config(fake, fmt.Sprintf(`
resource "bluestonepim_relation" "successor" {
  name          = "Superseded by"
  force_destroy = %t
}
`, forceDestroy))
	}

	runTest(t,
		resource.TestStep{
			Config: relation(false),
			Check: func(s *terraform.State) (err error) {
				id, err = stateValue(s, "bluestonepim_relation.successor", "id")
				fake.AddRelationConnection(id)
				return err
			},
		},
		resource.TestStep{
			Config:      relation(false),
			Destroy:     true,
			ExpectError: regexpMustCompile("force_destroy = true"),
		},
		resource.TestStep{
			Config: relation(true),
		},
		resource.TestStep{
			Config:  relation(true),
			Destroy: true,
			Check: func(*terraform.State) error {
				if fake.Relation(id) != nil {
					return fmt.Errorf("expected the relation to be deleted")
				}
				return nil
			},
		},
	)
}
