package relation

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
)

// GetRelationByID returns nil when the relation does not exist.
func GetRelationByID(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*Relation, diag.Diagnostic) {
	resp, err := client.FindOneWithResponse(ctx, id, &pim.FindOneParams{})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to read data", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil, nil
	}

	if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
		return nil, d
	}

	r := resp.JSON200
	return &Relation{
		Id:              types.StringPointerValue(r.Id),
		Name:            types.StringPointerValue(r.Name),
		Number:          nullIfEmpty(r.Number),
		Description:     nullIfEmpty(r.Description),
		ReverseName:     nullIfEmpty(r.ReverseName),
		Direction:       nullIfEmpty(r.Direction),
		QuantityEnabled: types.BoolPointerValue(r.QuantityEnabled),
	}, nil
}

func CreateRelation(ctx context.Context, client pim.ClientWithResponsesInterface, planned *Relation) (*Relation, diag.Diagnostic) {
	resp, err := client.CreateRelationWithResponse(ctx, &pim.CreateRelationParams{}, pim.CreateRelationJSONRequestBody{
		Name:            planned.Name.ValueString(),
		Number:          utils.KnownStringPointer(planned.Number),
		Description:     planned.Description.ValueStringPointer(),
		ReverseName:     planned.ReverseName.ValueStringPointer(),
		Direction:       utils.KnownStringPointer(planned.Direction),
		QuantityEnabled: knownBoolPointer(planned.QuantityEnabled),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Unable to create relation", err.Error())
	}

	if d := utils.AssertStatusCode(resp, http.StatusCreated); d != nil {
		return nil, d
	}

	return getExisting(ctx, client, resp.HTTPResponse.Header.Get("Resource-Id"))
}

func UpdateRelation(ctx context.Context, client pim.ClientWithResponsesInterface, current, planned *Relation) (*Relation, diag.Diagnostic) {
	id := current.Id.ValueString()

	// Partial update, so only the changed fields are sent
	body := pim.PartialUpdateRelationJSONRequestBody{}
	changed := false
	if !planned.Name.Equal(current.Name) {
		body.Name = &pim.PropertyUpdateString{Value: planned.Name.ValueStringPointer()}
		changed = true
	}
	if !planned.Description.Equal(current.Description) {
		body.Description = &pim.PropertyUpdateString{Value: planned.Description.ValueStringPointer()}
		changed = true
	}
	if !planned.ReverseName.Equal(current.ReverseName) {
		body.ReverseName = &pim.PropertyUpdateString{Value: planned.ReverseName.ValueStringPointer()}
		changed = true
	}
	if isKnown(planned.Number) && !planned.Number.Equal(current.Number) {
		body.Number = planned.Number.ValueStringPointer()
		changed = true
	}
	if isKnown(planned.QuantityEnabled) && !planned.QuantityEnabled.Equal(current.QuantityEnabled) {
		body.QuantityEnabled = planned.QuantityEnabled.ValueBoolPointer()
		changed = true
	}

	if changed {
		resp, err := client.PartialUpdateRelationWithResponse(ctx, id, &pim.PartialUpdateRelationParams{}, body)
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update relation", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	// The direction has a separate endpoint, the change is processed by a
	// background task which reconciles the product connections.
	directionChanged := isKnown(planned.Direction) && !planned.Direction.Equal(current.Direction)
	if directionChanged {
		resp, err := client.UpdateRelationDirectionWithResponse(ctx, id, pim.UpdateRelationDirectionJSONRequestBody{
			Direction: pim.RelationDirectionUpdateRequestDirection(planned.Direction.ValueString()),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Unable to update relation direction", err.Error())
		}

		if d := utils.AssertStatusCode(resp, http.StatusOK); d != nil {
			return nil, d
		}
	}

	result, d := getExisting(ctx, client, id)
	if d != nil {
		return nil, d
	}
	if directionChanged {
		// The background task may not have finished yet
		result.Direction = planned.Direction
	}
	return result, nil
}

// DeleteRelation deletes the relation. Without force the API refuses to
// delete a relation that is used by product connections; with force those
// connections are removed as well.
func DeleteRelation(ctx context.Context, client pim.ClientWithResponsesInterface, id string, force bool) diag.Diagnostic {
	resp, err := client.DeleteWithResponse(ctx, id, &pim.DeleteParams{Force: utils.Ref(force)})
	if err != nil {
		return diag.NewErrorDiagnostic("Unable to delete relation", err.Error())
	}

	if utils.IsNotFound(resp) {
		return nil
	}

	if resp.StatusCode() == http.StatusBadRequest && !force {
		d := utils.AssertStatusCode(resp, http.StatusNoContent)
		return diag.NewErrorDiagnostic("Unable to delete relation",
			d.Detail()+"\n\nThe relation is probably still used by product connections. Set `force_destroy = true` "+
				"to delete the relation together with all its product connections.")
	}

	return utils.AssertStatusCode(resp, http.StatusNoContent)
}

func getExisting(ctx context.Context, client pim.ClientWithResponsesInterface, id string) (*Relation, diag.Diagnostic) {
	result, d := GetRelationByID(ctx, client, id)
	if d == nil && result == nil {
		d = utils.NotFoundError("Relation", id)
	}
	return result, d
}

func nullIfEmpty(v *string) types.String {
	if v == nil || *v == "" {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func knownBoolPointer(v types.Bool) *bool {
	if !isKnown(v) {
		return nil
	}
	return v.ValueBoolPointer()
}

func isKnown(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}
