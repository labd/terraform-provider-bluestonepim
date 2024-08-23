package utils

import (
	"fmt"
	"github.com/labd/bluestonepim-go-sdk/notification_external"
	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type ProviderData struct {
	PimClient          *pim.ClientWithResponses
	NotificationClient *notification_external.ClientWithResponses
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
