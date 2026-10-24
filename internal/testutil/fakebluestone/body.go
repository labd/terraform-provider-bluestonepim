package fakebluestone

import "context"

type bodyKey struct{}

// The request body is consumed by the logging middleware, the handlers read
// the decoded body from the context.
func withBody(ctx context.Context, body map[string]any) context.Context {
	return context.WithValue(ctx, bodyKey{}, body)
}

func bodyFrom(ctx context.Context) map[string]any {
	body, _ := ctx.Value(bodyKey{}).(map[string]any)
	if body == nil {
		return map[string]any{}
	}
	return body
}

// propertyUpdate reads a PropertyUpdate* field: present returns the value (nil
// for an explicit null) and true.
func propertyUpdate(body map[string]any, field string) (any, bool) {
	wrapper, ok := body[field].(map[string]any)
	if !ok {
		return nil, false
	}
	return wrapper["value"], true
}
