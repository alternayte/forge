package schema

// OptionType represents the type of resource-level option.
type OptionType int

const (
	OptSoftDelete OptionType = iota
	OptAuditable
	OptTenantScoped
	OptSearchable
)

// String returns the string representation of the option type.
func (ot OptionType) String() string {
	switch ot {
	case OptSoftDelete:
		return "SoftDelete"
	case OptAuditable:
		return "Auditable"
	case OptTenantScoped:
		return "TenantScoped"
	case OptSearchable:
		return "Searchable"
	default:
		return "Unknown"
	}
}

// Option represents a resource-level option.
type Option struct {
	optionType OptionType
}

// Type returns the option type.
func (o *Option) Type() OptionType {
	return o.optionType
}

// schemaItem implements the SchemaItem interface.
func (o *Option) schemaItem() {}

// SoftDelete enables soft-delete functionality (deleted_at timestamp).
func SoftDelete() *Option {
	return &Option{optionType: OptSoftDelete}
}

// Auditable enables audit fields (created_by, updated_by, deleted_by).
func Auditable() *Option {
	return &Option{optionType: OptAuditable}
}

// TenantScoped enables multi-tenancy scoping (tenant_id foreign key).
func TenantScoped() *Option {
	return &Option{optionType: OptTenantScoped}
}

// Searchable enables full-text search indexing.
func Searchable() *Option {
	return &Option{optionType: OptSearchable}
}

// TimestampsItem represents the Timestamps() marker.
type TimestampsItem struct{}

// schemaItem implements the SchemaItem interface.
func (t *TimestampsItem) schemaItem() {}

// Timestamps signals that created_at and updated_at fields should be auto-added.
func Timestamps() *TimestampsItem {
	return &TimestampsItem{}
}
