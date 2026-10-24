// Package cla contains the shared logic of the Category Level Attribute (CLA)
// resources. Bluestone has separate endpoints per attribute definition type,
// so there is one resource per type, built on these helpers.
package cla

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// Flags are the propagation, lock and mandatory settings of a CLA.
type Flags struct {
	Mandatory types.Bool
	Propagate types.Bool
	Locked    types.Bool
}

// Get returns the CLA assigned on the category itself, or nil when the
// attribute is not assigned (or only inherited from a parent category).
func Get(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*pim.CategoryAttributeMetadataResponse, diag.Diagnostic) {
	response, err := client.ListAttributesAttachedToGivenNodeWithResponse(
		ctx, categoryId, &pim.ListAttributesAttachedToGivenNodeParams{})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(response) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(response, http.StatusOK); d != nil {
		return nil, d
	}

	if response.JSON200.Data == nil {
		return nil, nil
	}

	for _, attribute := range *response.JSON200.Data {
		if attribute.AttributeDefinitionId == nil || *attribute.AttributeDefinitionId != attributeId {
			continue
		}
		// Inherited attributes are managed on the category they are assigned on
		if attribute.AssignedOn == nil || *attribute.AssignedOn != categoryId {
			continue
		}
		return &attribute, nil
	}

	return nil, nil
}

// MustGet is Get for after a write, where the CLA has to exist.
func MustGet(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) (*pim.CategoryAttributeMetadataResponse, diag.Diagnostic) {
	result, d := Get(ctx, client, categoryId, attributeId)
	if d == nil && result == nil {
		d = diag.NewErrorDiagnostic("Category attribute not found",
			fmt.Sprintf("Attribute %s was not found on category %s", attributeId, categoryId))
	}
	return result, d
}

// FlagsFromResponse reports a flag as set only when this category switched it
// on. Flags switched on by a parent can only be changed on that parent.
func FlagsFromResponse(r *pim.CategoryAttributeMetadataResponse, categoryId string) Flags {
	setHere := func(v *string) types.Bool {
		return types.BoolValue(v != nil && *v == categoryId)
	}
	return Flags{
		Mandatory: setHere(r.MandatorySetOn),
		Propagate: setHere(r.CopySetOn),
		Locked:    setHere(r.LockedSetOn),
	}
}

// ApplyChangedFlags sets the changed flags on the request body. Returns
// whether anything changed.
func ApplyChangedFlags(current, planned Flags, mandatory, copyAttribute, locked **bool) bool {
	changed := false
	if planned.Mandatory.ValueBool() != current.Mandatory.ValueBool() {
		*mandatory = planned.Mandatory.ValueBoolPointer()
		changed = true
	}
	if planned.Propagate.ValueBool() != current.Propagate.ValueBool() {
		*copyAttribute = planned.Propagate.ValueBoolPointer()
		changed = true
	}
	if locked != nil && planned.Locked.ValueBool() != current.Locked.ValueBool() {
		*locked = planned.Locked.ValueBoolPointer()
		changed = true
	}
	return changed
}

// Patch updates the value and / or flags of a CLA. This is the only endpoint
// that sets the flags for every definition type.
func Patch(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string, body pim.UpdateNodeAttributeValueJSONRequestBody) diag.Diagnostic {
	response, err := client.UpdateNodeAttributeValueWithResponse(ctx, categoryId, attributeId, &pim.UpdateNodeAttributeValueParams{}, body)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to update category attribute", err.Error())
	}
	return utils.AssertStatusCode(response, http.StatusAccepted)
}

// UpdateFlags patches the flags when they changed.
func UpdateFlags(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string, current, planned Flags, supportsLocked bool) diag.Diagnostic {
	body := pim.UpdateNodeAttributeValueJSONRequestBody{}
	var locked **bool
	if supportsLocked {
		locked = &body.Locked
	}
	if !ApplyChangedFlags(current, planned, &body.Mandatory, &body.CopyAttribute, locked) {
		return nil
	}
	return Patch(ctx, client, categoryId, attributeId, body)
}

func Delete(ctx context.Context, client pim.ClientWithResponsesInterface, categoryId, attributeId string) diag.Diagnostic {
	response, err := client.DeleteAttributeFromNodeWithResponse(ctx, categoryId, attributeId)
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to remove attribute definition from category", err.Error())
	}

	if utils.IsNotFound(response) {
		return nil
	}

	return utils.AssertStatusCode(response, http.StatusNoContent)
}

// ImportState imports a CLA by `<category_id>/<attribute_definition_id>`.
func ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected `<category_id>/<attribute_definition_id>`, got %q", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("category_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("attribute_definition_id"), parts[1])...)
}

// ValidateLocked checks that a locked attribute is also mandatory. Locking an
// attribute makes it mandatory in Bluestone, so `mandatory = false` would
// never converge.
func ValidateLocked(locked, mandatory types.Bool, resp *resource.ValidateConfigResponse) {
	if locked.IsUnknown() || mandatory.IsUnknown() {
		return
	}

	if locked.ValueBool() && !mandatory.ValueBool() {
		resp.Diagnostics.AddAttributeError(path.Root("mandatory"), "Invalid attribute combination",
			"A locked attribute is always mandatory, set `mandatory = true` when `locked = true`.")
	}
}
