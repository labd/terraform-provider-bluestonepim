package utils

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func Ref[T any](s T) *T {
	return &s
}

func OptionalValueString(s basetypes.StringValue) *string {
	if s.IsNull() {
		return nil
	}
	return Ref(s.ValueString())
}

func NewStringValue(s *string) basetypes.StringValue {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}
