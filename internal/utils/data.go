package utils

import (
	"fmt"
	"github.com/labd/terraform-provider-bluestonepim/internal/sdk/notifications"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/labd/terraform-provider-bluestonepim/internal/sdk/pim"
)

type ProviderData struct {
	PimClient          *pim.ClientWithResponses
	NotificationClient *notifications.ClientWithResponses
}

func GetProviderData(data any) (*ProviderData, diag.Diagnostic) {
	d, ok := data.(*ProviderData)
	if !ok {
		return nil, diag.NewErrorDiagnostic(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *utils.ProviderData, got: %T. Please report this issue to the provider developers.", data),
		)
	}
	return d, nil
}
