package schema

// PermissionItem represents a resource-level permission rule.
// It specifies which roles are allowed to perform a given operation.
type PermissionItem struct {
	Operation string   // One of: "list", "read", "create", "update", "delete"
	Roles     []string // Roles allowed to perform the operation
}

// schemaItem implements the SchemaItem interface.
func (p *PermissionItem) schemaItem() {}

// Permission creates a resource-level permission rule restricting the given
// operation to the specified roles.
//
// Example: schema.Permission("list", "admin", "editor")
func Permission(operation string, roles ...string) *PermissionItem {
	return &PermissionItem{
		Operation: operation,
		Roles:     roles,
	}
}
