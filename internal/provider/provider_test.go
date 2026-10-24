package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/labd/terraform-provider-bluestonepim/internal/testutil/fakebluestone"
)

// The tests in this package run Terraform against the provider and an
// in-memory fake of the Bluestone API (see fakebluestone). After every apply
// the test framework also checks that a new plan is empty.

var testProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bluestonepim": providerserver.NewProtocol6WithError(New("test", false)()),
}

func runTest(t *testing.T, steps ...resource.TestStep) {
	t.Helper()
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProviderFactories,
		Steps:                    steps,
	})
}

// config prefixes the configuration with the provider block for the fake.
func config(fake *fakebluestone.Server, body string) string {
	return fake.ProviderConfig() + body
}

// stateValue reads an attribute from the state.
func stateValue(s *terraform.State, address, attribute string) (string, error) {
	rs, ok := s.RootModule().Resources[address]
	if !ok {
		return "", fmt.Errorf("resource %s not found in state", address)
	}
	value, ok := rs.Primary.Attributes[attribute]
	if !ok {
		return "", fmt.Errorf("attribute %s not found on %s", attribute, address)
	}
	return value, nil
}

func regexpMustCompile(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}
