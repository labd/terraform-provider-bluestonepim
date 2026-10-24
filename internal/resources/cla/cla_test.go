package cla

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestApplyChangedFlags(t *testing.T) {
	current := Flags{Mandatory: types.BoolValue(true), Propagate: types.BoolValue(false), Locked: types.BoolValue(false)}
	planned := Flags{Mandatory: types.BoolValue(true), Propagate: types.BoolValue(true), Locked: types.BoolValue(true)}

	var mandatory, propagate, locked *bool
	if !ApplyChangedFlags(current, planned, &mandatory, &propagate, &locked) {
		t.Fatal("expected changes")
	}
	if mandatory != nil {
		t.Errorf("unchanged mandatory should not be sent")
	}
	if propagate == nil || !*propagate || locked == nil || !*locked {
		t.Errorf("expected propagate and locked to be sent as true")
	}

	// Without lock support the locked flag is never sent
	mandatory, propagate = nil, nil
	ApplyChangedFlags(current, planned, &mandatory, &propagate, nil)
	if propagate == nil {
		t.Errorf("expected propagate to be sent")
	}

	if ApplyChangedFlags(current, current, &mandatory, &propagate, &locked) {
		t.Errorf("expected no changes")
	}
}
