package schema

// JobRef identifies a River job to enqueue after a resource lifecycle event.
// Kind maps to the job's worker kind string; Queue names the River queue.
//
// Example usage:
//
//	schema.WithHooks(schema.Hooks{
//	    AfterCreate: []schema.JobRef{
//	        {Kind: "notify_new_product", Queue: "notifications"},
//	    },
//	})
type JobRef struct {
	Kind  string // River job worker kind (e.g., "notify_new_product")
	Queue string // River queue name (e.g., "notifications")
}

// Hooks declares lifecycle job enqueueing for a resource.
// AfterCreate jobs are enqueued in the same transaction as the CREATE operation.
// AfterUpdate jobs are enqueued in the same transaction as the UPDATE operation.
type Hooks struct {
	AfterCreate []JobRef // Jobs to enqueue after resource creation
	AfterUpdate []JobRef // Jobs to enqueue after resource update
}

// HooksItem wraps Hooks and implements the SchemaItem interface so it can
// be passed as a variadic argument to schema.Define().
type HooksItem struct {
	Hooks Hooks
}

// schemaItem implements the SchemaItem interface.
func (h *HooksItem) schemaItem() {}

// WithHooks creates a HooksItem that declares lifecycle River job enqueueing
// for the enclosing resource definition.
//
// Example:
//
//	schema.Define("Product",
//	    schema.WithHooks(schema.Hooks{
//	        AfterCreate: []schema.JobRef{
//	            {Kind: "notify_new_product", Queue: "notifications"},
//	        },
//	    }),
//	)
func WithHooks(h Hooks) *HooksItem {
	return &HooksItem{Hooks: h}
}
