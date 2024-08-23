package utils

import (
	"fmt"
	"github.com/labd/bluestonepim-go-sdk/pim"
	"reflect"

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
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	field := val.FieldByName(fmt.Sprintf("JSON%d", response.StatusCode()))
	if field.IsValid() && !field.IsNil() {
		return field.Interface().(*pim.ErrorResponse)
	}
	return nil
}
