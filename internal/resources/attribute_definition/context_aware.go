package attribute_definition

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// ContextAwareSchema is the `context_aware` attribute of all attribute
// definition resources.
func ContextAwareSchema() schema.BoolAttribute {
	return schema.BoolAttribute{
		MarkdownDescription: "Whether the attribute can have a different value per context (language or " +
			"channel). When `false` the attribute is global: it has one value, shared by all contexts. Uses the " +
			"Bluestone default when not set. Cannot be changed after creation.",
		Optional: true,
		Computed: true,
		PlanModifiers: []planmodifier.Bool{
			boolplanmodifier.UseStateForUnknown(),
			contextAwareImmutable{},
		},
	}
}

// contextAwareImmutable rejects changing `context_aware` of an existing
// definition. Bluestone switches it with a background job that cleans up the
// related values, so switching from context-aware to global can remove the
// values of the other contexts.
type contextAwareImmutable struct{}

func (m contextAwareImmutable) Description(_ context.Context) string {
	return "context_aware of an existing attribute definition cannot be changed."
}

func (m contextAwareImmutable) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m contextAwareImmutable) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.State.Raw.IsNull() || req.StateValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	if !req.PlanValue.Equal(req.StateValue) {
		resp.Diagnostics.AddAttributeError(req.Path, "context_aware cannot be changed",
			fmt.Sprintf("`context_aware` of an existing attribute definition cannot be changed from %t to %t. "+
				"Bluestone switches it with a background job that cleans up related values, which can remove the "+
				"values of other contexts. Switch it in Bluestone if that is intended, and update the configuration "+
				"afterwards.", req.StateValue.ValueBool(), req.PlanValue.ValueBool()))
	}
}
