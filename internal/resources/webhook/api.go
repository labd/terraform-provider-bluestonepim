package webhook

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/terraform-provider-bluestonepim/internal/sdk/notifications"
	"slices"
)

const ResourceIdHeader = "Resource-Id"

func GetWebhookByID(
	ctx context.Context,
	client *notifications.ClientWithResponses,
	id string,
) (*Webhook, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	webhookRes, err := client.GetWithResponse(ctx, id)
	if err != nil {
		diags.AddError("Failed fetching webhook", err.Error())
		return nil, diags
	}
	if webhookRes.StatusCode() != 200 {
		diags.AddError(
			"Failed fetching webhook",
			fmt.Sprintf("unexpected status code %d", webhookRes.StatusCode()),
		)
		return nil, diags
	}

	subscriptionRes, err := client.FindWebhookWithResponse(ctx, id)
	if err != nil {
		diags.AddError("Failed fetching webhook subscriptions", err.Error())
		return nil, diags
	}
	if subscriptionRes.StatusCode() != 200 {
		diags.AddError(
			"Failed fetching webhook subscriptions",
			fmt.Sprintf("unexpected status code %d", subscriptionRes.StatusCode()),
		)
		return nil, diags
	}

	eventTypes, diags := types.ListValueFrom(ctx, types.StringType, subscriptionRes.JSON200.EventTypes)
	if diags.HasError() {
		return nil, diags
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
	client *notifications.ClientWithResponses,
	current *Webhook,
) (*Webhook, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	webhookRes, err := client.Create(ctx, notifications.CreateJSONRequestBody{
		Secret: current.Secret.ValueString(),
		Url:    current.URL.ValueString(),
		Active: current.Active.ValueBool(),
	})
	if err != nil {
		diags.AddError("Failed creating webhook", err.Error())
		return nil, diags
	}

	if webhookRes.StatusCode != 201 {
		diags.AddError("Failed creating webhook", fmt.Sprintf("unexpected status code %d", webhookRes.StatusCode))
		return nil, diags
	}

	id := webhookRes.Header.Get(ResourceIdHeader)
	if id == "" {
		diags.AddError(
			"Failed creating webhook",
			fmt.Sprintf("missing resource id. Expected header '%s' to be set", ResourceIdHeader),
		)
		return nil, diags
	}

	var eventTypes []notifications.WebhookEventTypeListRequestEventTypes
	diags = current.EventTypes.ElementsAs(ctx, &eventTypes, false)
	if diags.HasError() {
		return nil, diags
	}

	subscriptionRes, err := client.Subscribe(ctx, id, notifications.SubscribeJSONRequestBody{EventTypes: eventTypes})
	if err != nil {
		diags.AddError("Failed adding subscriptions to webhook", err.Error())
		return nil, diags
	}
	if subscriptionRes.StatusCode != 200 {
		diags.AddError(
			"Failed adding subscriptions to webhook",
			fmt.Sprintf("unexpected status code %d", subscriptionRes.StatusCode),
		)
		return nil, diags
	}

	return GetWebhookByID(ctx, client, id)
}

func UpdateWebhookById(
	ctx context.Context,
	client *notifications.ClientWithResponses,
	id string,
	current *Webhook,
	planned *Webhook,
) (*Webhook, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if !(current.ID.Equal(planned.ID) && current.Secret.Equal(planned.Secret) && current.URL.Equal(planned.URL) && current.Active.Equal(planned.Active)) {
		updateRes, err := client.Update(ctx, id, notifications.UpdateJSONRequestBody{
			Secret: planned.Secret.ValueStringPointer(),
			Url:    planned.URL.ValueStringPointer(),
			Active: planned.Active.ValueBoolPointer(),
		})
		if err != nil {
			diags.AddError("Failed updating webhook", err.Error())
			return nil, diags
		}
		if updateRes.StatusCode != 200 {
			diags.AddError(
				"Failed updating webhook",
				fmt.Sprintf("unexpected status code %d", updateRes.StatusCode),
			)
			return nil, diags
		}
	}

	if !current.EventTypes.Equal(planned.EventTypes) {
		var currentEventTypes []notifications.WebhookEventTypeListRequestEventTypes
		diags = current.EventTypes.ElementsAs(ctx, &currentEventTypes, false)
		if diags.HasError() {
			return nil, diags
		}

		var plannedEventTypes []notifications.WebhookEventTypeListRequestEventTypes
		diags = planned.EventTypes.ElementsAs(ctx, &plannedEventTypes, false)
		if diags.HasError() {
			return nil, diags
		}

		var unsubscribeEventTypes []notifications.WebhookEventTypeListRequestEventTypes
		for _, currentEventType := range currentEventTypes {
			if slices.Contains(plannedEventTypes, currentEventType) {
				continue
			}
			unsubscribeEventTypes = append(unsubscribeEventTypes, currentEventType)
		}
		if len(unsubscribeEventTypes) > 0 {
			res, err := client.UnsubscribeWithResponse(ctx, id, notifications.UnsubscribeJSONRequestBody{EventTypes: unsubscribeEventTypes})
			if err != nil {
				diags.AddError("Failed removing subscriptions from webhook", err.Error())
				return nil, diags
			}
			if res.StatusCode() == 400 {
				diags.AddError("Failed removing subscriptions from webhook", *res.JSON400.Error)
				return nil, diags
			}
			if res.StatusCode() != 200 {
				diags.AddError(
					"Failed removing subscriptions from webhook",
					fmt.Sprintf("unexpected status code %d", res.StatusCode()),
				)
				return nil, diags
			}
		}

		var subscribeEventTypes []notifications.WebhookEventTypeListRequestEventTypes
		for _, plannedEventType := range plannedEventTypes {
			if slices.Contains(currentEventTypes, plannedEventType) {
				continue
			}
			subscribeEventTypes = append(subscribeEventTypes, plannedEventType)
		}

		if len(subscribeEventTypes) > 0 {
			res, err := client.SubscribeWithResponse(ctx, id, notifications.SubscribeJSONRequestBody{EventTypes: subscribeEventTypes})
			if err != nil {
				diags.AddError("Failed adding subscriptions to webhook", err.Error())
				return nil, diags
			}
			if res.StatusCode() == 400 {
				diags.AddError("Failed adding subscriptions to webhook", *res.JSON400.Error)
				return nil, diags
			}
			if res.StatusCode() != 200 {
				diags.AddError(
					"Failed adding subscriptions to webhook",
					fmt.Sprintf("unexpected status code %d", res.StatusCode()),
				)
				return nil, diags
			}
		}

	}

	return GetWebhookByID(ctx, client, id)
}

func DeleteWebhookByID(
	ctx context.Context,
	client *notifications.ClientWithResponses,
	id string,
) diag.Diagnostics {
	diags := diag.Diagnostics{}
	res, err := client.Delete(ctx, id)
	if err != nil {
		diags.AddError("Failed deleting webhook", err.Error())
		return diags
	}
	if res.StatusCode != 200 {
		diags.AddError(
			"Failed deleting webhook",
			fmt.Sprintf("unexpected status code %d", res.StatusCode),
		)
		return diags
	}

	return diags
}
