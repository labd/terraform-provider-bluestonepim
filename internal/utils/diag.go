package utils

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/labd/bluestonepim-go-sdk/pim"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type Response interface {
	StatusCode() int
}

func AssertStatusCode(response Response, statusCode int) diag.Diagnostic {
	if response.StatusCode() == statusCode {
		return nil
	}

	if response.StatusCode() >= 400 && response.StatusCode() < 500 {
		e := getErrorResponse(response)
		if e != nil {
			return diag.NewErrorDiagnostic(
				fmt.Sprintf("HTTP %d error", response.StatusCode()), *e.Error)
		}
	}

	return diag.NewErrorDiagnostic("Unexpected status code", fmt.Sprintf("Expected %d, got %d", statusCode, response.StatusCode()))
}

func getErrorResponse(response Response) *pim.ErrorResponse {
	val := reflect.ValueOf(response)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	field := val.FieldByName(fmt.Sprintf("JSON%d", response.StatusCode()))
	if field.IsValid() && !field.IsNil() {
		return field.Interface().(*pim.ErrorResponse)
	}
	return nil
}

// AssertSuccessStatusCode accepts any 2xx status code. Used for endpoints where
// the exact success status code is not documented in the SDK.
func AssertSuccessStatusCode(response Response) diag.Diagnostic {
	if response.StatusCode() >= 200 && response.StatusCode() < 300 {
		return nil
	}
	return AssertStatusCode(response, http.StatusOK)
}

func IsNotFound(response Response) bool {
	return response.StatusCode() == http.StatusNotFound
}

// NotFoundError is returned when a resource that was just written, or that is
// looked up by a data source, does not exist.
func NotFoundError(kind, id string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(kind+" not found", fmt.Sprintf("%s %s was not found", kind, id))
}
