package utils

import (
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Reorder moves the desired items to the start of the current order, in the
// desired order. Items not in desired keep their relative order after them.
// The API only moves one item at a time to a (0-based) position, so move is
// called for each item that is not in place yet.
func Reorder(current, desired []string, move func(id string, position int) diag.Diagnostic) diag.Diagnostic {
	order := slices.Clone(current)
	for position, id := range desired {
		index := slices.Index(order, id)
		if index == -1 {
			return diag.NewErrorDiagnostic("Unknown item", fmt.Sprintf("%s does not exist", id))
		}
		if index == position {
			continue
		}

		if d := move(id, position); d != nil {
			return d
		}
		order = slices.Insert(slices.Delete(order, index, index+1), position, id)
	}
	return nil
}

// LeadingItems returns the first n items of the current order, which are the
// items that should be in the desired order after Reorder. Used to read back
// the order: when other items are in between, the state shows them and the
// next plan moves the desired items back to the start.
func LeadingItems(current []string, n int) []string {
	if n > len(current) {
		n = len(current)
	}
	return slices.Clone(current[:n])
}
