package parser

// ResourceIR represents a top-level parsed resource definition.
// This is the domain model IR that the parser produces and the code generator consumes.
type ResourceIR struct {
	Name          string              // Resource name (e.g., "Product")
	Fields        []FieldIR           // Parsed fields
	Relationships []RelationshipIR    // Parsed relationships
	Options       ResourceOptionsIR   // Resource-level options
	HasTimestamps bool                // Whether schema.Timestamps() was called
	SourceFile    string              // Original file path for error reporting
	SourceLine    int                 // Line number of schema.Define() call
}

// FieldIR represents a parsed field definition.
type FieldIR struct {
	Name       string        // Field name
	Type       string        // Field type as string (e.g., "String", "UUID", "Enum")
	Modifiers  []ModifierIR  // Applied modifiers
	EnumValues []string      // For Enum fields, the allowed values
	SourceLine int           // Line number in source file
}

// ModifierIR represents a parsed modifier on a field.
type ModifierIR struct {
	Type       string      // Modifier name (e.g., "Required", "MaxLen", "Default")
	Value      interface{} // Modifier argument if any (int for MaxLen, string for Default, etc.)
	SourceLine int         // Line number in source file
}

// RelationshipIR represents a parsed relationship definition.
type RelationshipIR struct {
	Name       string // Relationship name
	Type       string // "BelongsTo", "HasMany", "HasOne", "ManyToMany"
	Table      string // Related table name
	OnDelete   string // Cascade action (empty string = default)
	Optional   bool   // Whether relationship is optional
	Eager      bool   // Whether to eager-load this relationship
	SourceLine int    // Line number in source file
}

// PermissionsIR maps operation names to the roles allowed to perform them.
// Example: {"list": ["admin", "editor"], "delete": ["admin"]}
type PermissionsIR map[string][]string

// ResourceOptionsIR represents resource-level options.
type ResourceOptionsIR struct {
	SoftDelete   bool          // Enable soft delete
	Auditable    bool          // Enable audit logging
	TenantScoped bool          // Enable multi-tenancy scoping
	Searchable   bool          // Enable full-text search
	Permissions  PermissionsIR // Role-based permission rules per operation
}

// ParseResult represents the output of parsing a directory of schema files.
// It collects all resources and all errors encountered (not just the first one).
type ParseResult struct {
	Resources []ResourceIR // All successfully parsed resources
	Errors    []error      // All parse errors collected during parsing
}
