package utils

import "github.com/hashicorp/terraform-plugin-framework/types"

func Ref[T any](s T) *T {
	return &s
}

// KnownStringPointer returns nil for null and unknown values. Unknown values
// would otherwise be sent to the API as an empty string.
func KnownStringPointer(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return v.ValueStringPointer()
}
