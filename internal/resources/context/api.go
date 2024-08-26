package context

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/global_settings"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
	"net/http"
)

const ResourceIdHeader = "Resource-Id"

func GetContextByID(
	ctx context.Context,
	client global_settings.ClientWithResponsesInterface,
	id string,
) (*Context, diag.Diagnostic) {
	contextRes, err := client.GetWithResponse(ctx, id)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed fetching context", err.Error())
	}
	if d := utils.AssertStatusCode(contextRes, http.StatusOK); d != nil {
		return nil, d
	}
	return &Context{
		ID:         types.StringValue(contextRes.JSON200.Id),
		Name:       types.StringValue(contextRes.JSON200.Name),
		Locale:     types.StringValue(contextRes.JSON200.Locale),
		FallbackID: types.StringPointerValue(contextRes.JSON200.Fallback),
	}, nil
}

func CreateContext(
	ctx context.Context,
	//We need the implementation to do Create instead of CreateWithResponse, as the API returns an invalid response when creating a context
	client *global_settings.ClientWithResponses,
	current *Context,
) (*Context, diag.Diagnostic) {
	contextRes, err := client.Create(ctx, global_settings.CreateJSONRequestBody{
		Fallback: current.FallbackID.ValueStringPointer(),
		Locale:   current.Locale.ValueString(),
		Name:     current.Name.ValueString(),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed creating context", err.Error())
	}

	if contextRes.StatusCode != http.StatusCreated {
		return nil, diag.NewErrorDiagnostic("Failed creating context", fmt.Sprintf("unexpected status code %d", contextRes.StatusCode))
	}

	//Workaround to find the id for the newly created context based on locale
	res, err := client.FindWithResponse(ctx, nil)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed creating context", err.Error())
	}

	if d := utils.AssertStatusCode(res, http.StatusOK); d != nil {
		return nil, d
	}

	var foundContext *global_settings.ContextResponseDto
	for _, d := range res.JSON200.Data {
		if d.Locale == current.Locale.ValueString() {
			foundContext = &d
			break
		}
	}
	if foundContext == nil {
		return nil, diag.NewErrorDiagnostic("Failed finding context in list", fmt.Sprintf("Failed finding context with locale %s in list", current.Locale.ValueString()))
	}

	return GetContextByID(ctx, client, foundContext.Id)
}

func UpdateContextById(
	ctx context.Context,
	client global_settings.ClientWithResponsesInterface,
	id string,
	current *Context,
	planned *Context,
) (*Context, diag.Diagnostic) {
	if !(current.Locale.Equal(planned.Locale) && current.FallbackID.Equal(planned.FallbackID) && current.Name.Equal(planned.Name)) {
		updateRes, err := client.UpdateWithResponse(ctx, id, global_settings.UpdateJSONRequestBody{
			Fallback: planned.FallbackID.ValueStringPointer(),
			Locale:   planned.Locale.ValueString(),
			Name:     planned.Name.ValueString(),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Failed updating context", err.Error())
		}
		if d := utils.AssertStatusCode(updateRes, http.StatusNoContent); d != nil {
			return nil, d
		}
	}

	return GetContextByID(ctx, client, id)
}

func DeleteContextByID(
	ctx context.Context,
	client global_settings.ClientWithResponsesInterface,
	id string,
) diag.Diagnostic {
	res, err := client.ArchiveWithResponse(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Failed archiving context", err.Error())
	}
	if d := utils.AssertStatusCode(res, http.StatusNoContent); d != nil {
		return d
	}

	return nil
}
