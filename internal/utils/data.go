package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/labd/bluestone-pim-go-sdk/pim"
)

type ProviderData struct {
	Client *pim.ClientWithResponses
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
