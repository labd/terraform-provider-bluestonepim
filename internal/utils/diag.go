package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type Response interface {
	StatusCode() int
}

func AssertStatusCode(response Response, statusCode int) diag.Diagnostic {
	if response.StatusCode() == statusCode {
		return nil
	}
	return diag.NewErrorDiagnostic("Unexpected status code", fmt.Sprintf("Expected %s, got %d", statusCode, response.StatusCode()))
}
