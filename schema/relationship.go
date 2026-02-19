package schema

// RelationType represents the type of relationship between resources.
type RelationType int

const (
	RelBelongsTo RelationType = iota
	RelHasMany
	RelHasOne
	RelManyToMany
)

// String returns the string representation of the relationship type.
func (rt RelationType) String() string {
	switch rt {
	case RelBelongsTo:
		return "BelongsTo"
	case RelHasMany:
		return "HasMany"
	case RelHasOne:
		return "HasOne"
	case RelManyToMany:
		return "ManyToMany"
	default:
		return "Unknown"
	}
}

// Relationship represents a relationship between resources.
type Relationship struct {
	name      string
	relType   RelationType
	table     string
	onDelete  OnDeleteAction
	isOptional bool
}

// Name returns the relationship name.
func (r *Relationship) Name() string {
	return r.name
}

// RelType returns the relationship type.
func (r *Relationship) RelType() RelationType {
	return r.relType
}

// Table returns the related table name.
func (r *Relationship) Table() string {
	return r.table
}

// OnDeleteAction returns the on delete action.
func (r *Relationship) OnDeleteAction() OnDeleteAction {
	return r.onDelete
}

// IsOptional returns whether the relationship is optional.
func (r *Relationship) IsOptional() bool {
	return r.isOptional
}

// schemaItem implements the SchemaItem interface.
func (r *Relationship) schemaItem() {}

// Optional marks the relationship as optional (nullable foreign key).
func (r *Relationship) Optional() *Relationship {
	r.isOptional = true
	return r
}

// OnDelete sets the action to take when the referenced record is deleted.
func (r *Relationship) OnDelete(action OnDeleteAction) *Relationship {
	r.onDelete = action
	return r
}

// BelongsTo creates a belongs-to relationship.
func BelongsTo(name, table string) *Relationship {
	return &Relationship{
		name:     name,
		relType:  RelBelongsTo,
		table:    table,
		onDelete: NoAction, // Default
	}
}

// HasMany creates a has-many relationship.
func HasMany(name, table string) *Relationship {
	return &Relationship{
		name:     name,
		relType:  RelHasMany,
		table:    table,
		onDelete: NoAction, // Default
	}
}

// HasOne creates a has-one relationship.
func HasOne(name, table string) *Relationship {
	return &Relationship{
		name:     name,
		relType:  RelHasOne,
		table:    table,
		onDelete: NoAction, // Default
	}
}

// ManyToMany creates a many-to-many relationship.
func ManyToMany(name, table string) *Relationship {
	return &Relationship{
		name:     name,
		relType:  RelManyToMany,
		table:    table,
		onDelete: NoAction, // Default
	}
}
