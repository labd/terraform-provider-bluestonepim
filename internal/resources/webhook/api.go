package webhook

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/bluestonepim-go-sdk/notification_external"
	"github.com/labd/terraform-provider-bluestonepim/internal/utils"
	"net/http"
	"slices"
)

const ResourceIdHeader = "Resource-Id"

func GetWebhookByID(
	ctx context.Context,
	client *notification_external.ClientWithResponses,
	id string,
) (*Webhook, diag.Diagnostic) {
	webhookRes, err := client.GetWithResponse(ctx, id)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed fetching webhook", err.Error())
	}
	if d := utils.AssertStatusCode(webhookRes, http.StatusOK); d != nil {
		return nil, d
	}

	subscriptionRes, err := client.FindWebhookWithResponse(ctx, id)
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed fetching webhook subscriptions", err.Error())
	}
	if d := utils.AssertStatusCode(subscriptionRes, http.StatusOK); d != nil {
		return nil, d
	}

	eventTypes, diags := types.ListValueFrom(ctx, types.StringType, subscriptionRes.JSON200.EventTypes)
	if diags.HasError() {
		//Return the first error, but there might be more
		return nil, diags.Errors()[0]
	}

	webhook := &Webhook{
		ID:         types.StringValue(webhookRes.JSON200.Id),
		Secret:     types.StringValue(webhookRes.JSON200.Secret),
		URL:        types.StringValue(webhookRes.JSON200.Url),
		Active:     types.BoolValue(webhookRes.JSON200.Active),
		EventTypes: eventTypes,
	}

	return webhook, nil
}

func CreateWebhook(
	ctx context.Context,
	client *notification_external.ClientWithResponses,
	current *Webhook,
) (*Webhook, diag.Diagnostic) {
	webhookRes, err := client.Create(ctx, notification_external.CreateJSONRequestBody{
		Secret: current.Secret.ValueString(),
		Url:    current.URL.ValueString(),
		Active: current.Active.ValueBool(),
	})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed creating webhook", err.Error())
	}

	if webhookRes.StatusCode != http.StatusCreated {
		return nil, diag.NewErrorDiagnostic("Failed creating webhook", fmt.Sprintf("Expected status code %d, got %d", http.StatusCreated, webhookRes.StatusCode))
	}

	id := webhookRes.Header.Get(ResourceIdHeader)
	if id == "" {
		return nil, diag.NewErrorDiagnostic(
			"Failed creating webhook",
			fmt.Sprintf("missing resource id. Expected header '%s' to be set", ResourceIdHeader),
		)
	}

	var eventTypes []notification_external.WebhookEventTypeListRequestEventTypes
	diags := current.EventTypes.ElementsAs(ctx, &eventTypes, false)
	if diags.HasError() {
		return nil, diags.Errors()[0]
	}

	subscriptionRes, err := client.SubscribeWithResponse(ctx, id, notification_external.SubscribeJSONRequestBody{EventTypes: eventTypes})
	if err != nil {
		return nil, diag.NewErrorDiagnostic("Failed adding subscriptions to webhook", err.Error())
	}
	if d := utils.AssertStatusCode(subscriptionRes, http.StatusOK); d != nil {
		return nil, d
	}

	return GetWebhookByID(ctx, client, id)
}

func UpdateWebhookById(
	ctx context.Context,
	client *notification_external.ClientWithResponses,
	id string,
	current *Webhook,
	planned *Webhook,
) (*Webhook, diag.Diagnostic) {
	if !(current.ID.Equal(planned.ID) && current.Secret.Equal(planned.Secret) && current.URL.Equal(planned.URL) && current.Active.Equal(planned.Active)) {
		updateRes, err := client.Update(ctx, id, notification_external.UpdateJSONRequestBody{
			Secret: planned.Secret.ValueStringPointer(),
			Url:    planned.URL.ValueStringPointer(),
			Active: planned.Active.ValueBoolPointer(),
		})
		if err != nil {
			return nil, diag.NewErrorDiagnostic("Failed updating webhook", err.Error())
		}
		if updateRes.StatusCode != http.StatusOK {
			return nil, diag.NewErrorDiagnostic("Failed updating webhook", fmt.Sprintf("Expected status code %d, got %d", http.StatusNoContent, updateRes.StatusCode))
		}
	}

	if !current.EventTypes.Equal(planned.EventTypes) {
		var currentEventTypes []notification_external.WebhookEventTypeListRequestEventTypes
		diags := current.EventTypes.ElementsAs(ctx, &currentEventTypes, false)
		if diags.HasError() {
			return nil, diags.Errors()[0]
		}

		var plannedEventTypes []notification_external.WebhookEventTypeListRequestEventTypes
		diags = planned.EventTypes.ElementsAs(ctx, &plannedEventTypes, false)
		if diags.HasError() {
			return nil, diags.Errors()[0]
		}

		var unsubscribeEventTypes []notification_external.WebhookEventTypeListRequestEventTypes
		for _, currentEventType := range currentEventTypes {
			if slices.Contains(plannedEventTypes, currentEventType) {
				continue
			}
			unsubscribeEventTypes = append(unsubscribeEventTypes, currentEventType)
		}
		if len(unsubscribeEventTypes) > 0 {
			res, err := client.UnsubscribeWithResponse(ctx, id, notification_external.UnsubscribeJSONRequestBody{EventTypes: unsubscribeEventTypes})
			if err != nil {
				return nil, diag.NewErrorDiagnostic("Failed removing subscriptions from webhook", err.Error())
			}
			if d := utils.AssertStatusCode(res, http.StatusOK); d != nil {
				return nil, d
			}
		}

		var subscribeEventTypes []notification_external.WebhookEventTypeListRequestEventTypes
		for _, plannedEventType := range plannedEventTypes {
			if slices.Contains(currentEventTypes, plannedEventType) {
				continue
			}
			subscribeEventTypes = append(subscribeEventTypes, plannedEventType)
		}

		if len(subscribeEventTypes) > 0 {
			res, err := client.SubscribeWithResponse(ctx, id, notification_external.SubscribeJSONRequestBody{EventTypes: subscribeEventTypes})
			if err != nil {
				return nil, diag.NewErrorDiagnostic("Failed adding subscriptions to webhook", err.Error())
			}
			if d := utils.AssertStatusCode(res, http.StatusOK); d != nil {
				return nil, d
			}
		}

	}

	return GetWebhookByID(ctx, client, id)
}

func DeleteWebhookByID(
	ctx context.Context,
	client *notification_external.ClientWithResponses,
	id string,
) diag.Diagnostic {
	res, err := client.Delete(ctx, id)
	if err != nil {
		return diag.NewErrorDiagnostic("Failed deleting webhook", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		return diag.NewErrorDiagnostic("Failed deleting webhook", fmt.Sprintf("Expected status code %d, got %d", http.StatusOK, res.StatusCode))
	}

	return nil
}
