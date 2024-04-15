package utils

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type LogTransport struct {
	transport http.RoundTripper
}

var DebugTransport = &LogTransport{
	transport: http.DefaultTransport,
}

func (c *LogTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if os.Getenv("BSP_DEBUG") != "" {
		logRequest(request.Context(), request)
	}
	response, err := c.transport.RoundTrip(request)
	if os.Getenv("BSP_DEBUG") != "" {
		logResponse(request.Context(), response, err)
	}
	return response, err
}

const logRequestTemplate = `DEBUG:
---[ REQUEST ]--------------------------------------------------------
%s
----------------------------------------------------------------------
`

const logResponseTemplate = `DEBUG:
---[ RESPONSE ]-------------------------------------------------------
%s
----------------------------------------------------------------------
`

func logRequest(ctx context.Context, r *http.Request) {
	body, err := httputil.DumpRequestOut(r, true)
	if err != nil {
		return
	}
	tflog.Info(ctx, fmt.Sprintf(logRequestTemplate, body))
	log.Printf(logRequestTemplate, body)
}

func logResponse(ctx context.Context, r *http.Response, err error) {
	if err != nil {
		tflog.Info(ctx, fmt.Sprintf(logResponseTemplate, err))
		log.Printf(logResponseTemplate, err)
		return
	}
	body, err := httputil.DumpResponse(r, true)
	if err != nil {
		return
	}
	tflog.Info(ctx, fmt.Sprintf(logResponseTemplate, body))
	log.Printf(logResponseTemplate, body)
}
