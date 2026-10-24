package utils

import (
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// simulate applies the moves like the API does: remove and insert at position.
func simulate(t *testing.T, current, desired []string) ([]string, int) {
	t.Helper()
	order := slices.Clone(current)
	moves := 0
	d := Reorder(current, desired, func(id string, position int) diag.Diagnostic {
		index := slices.Index(order, id)
		order = slices.Insert(slices.Delete(order, index, index+1), position, id)
		moves++
		return nil
	})
	if d != nil {
		t.Fatal(d.Detail())
	}
	return order, moves
}

func TestReorder(t *testing.T) {
	order, moves := simulate(t, []string{"a", "b", "c", "d"}, []string{"c", "a"})
	if !slices.Equal(order, []string{"c", "a", "b", "d"}) {
		t.Errorf("unexpected order %v", order)
	}
	// Moving c to the front puts a in place as well
	if moves != 1 {
		t.Errorf("expected 1 move, got %d", moves)
	}

	order, moves = simulate(t, []string{"a", "b", "c", "d"}, []string{"d", "c", "b", "a"})
	if !slices.Equal(order, []string{"d", "c", "b", "a"}) || moves != 3 {
		t.Errorf("unexpected order %v after %d moves", order, moves)
	}
}

func TestReorderAlreadyInOrder(t *testing.T) {
	_, moves := simulate(t, []string{"a", "b", "c"}, []string{"a", "b"})
	if moves != 0 {
		t.Errorf("expected no moves, got %d", moves)
	}
}

func TestReorderUnknownItem(t *testing.T) {
	d := Reorder([]string{"a"}, []string{"x"}, func(string, int) diag.Diagnostic { return nil })
	if d == nil {
		t.Error("expected an error for an unknown item")
	}
}

func TestLeadingItems(t *testing.T) {
	if got := LeadingItems([]string{"a", "x", "b"}, 2); !slices.Equal(got, []string{"a", "x"}) {
		t.Errorf("unexpected %v", got)
	}
	if got := LeadingItems([]string{"a"}, 3); !slices.Equal(got, []string{"a"}) {
		t.Errorf("unexpected %v", got)
	}
}
