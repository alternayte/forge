package schema

// Definition represents a complete resource schema definition.
type Definition struct {
	name          string
	fields        []*Field
	relationships []*Relationship
	options       []*Option
	hasTimestamps bool
}

// Name returns the resource name.
func (d *Definition) Name() string {
	return d.name
}

// Fields returns all fields in the definition.
func (d *Definition) Fields() []*Field {
	return d.fields
}

// Relationships returns all relationships in the definition.
func (d *Definition) Relationships() []*Relationship {
	return d.relationships
}

// Options returns all options in the definition.
func (d *Definition) Options() []*Option {
	return d.options
}

// HasTimestamps returns whether the definition includes timestamps.
func (d *Definition) HasTimestamps() bool {
	return d.hasTimestamps
}

// Define creates a new schema definition with the given name and items.
// Items can include fields, relationships, options, and the timestamps marker.
// All items implement the SchemaItem interface.
//
// This function exists so developer code compiles, but the AST parser extracts
// definitions statically — Define() is never called at parse time.
func Define(name string, items ...SchemaItem) *Definition {
	def := &Definition{
		name:          name,
		fields:        []*Field{},
		relationships: []*Relationship{},
		options:       []*Option{},
		hasTimestamps: false,
	}

	for _, item := range items {
		switch v := item.(type) {
		case *Field:
			def.fields = append(def.fields, v)
		case *Relationship:
			def.relationships = append(def.relationships, v)
		case *Option:
			def.options = append(def.options, v)
		case *TimestampsItem:
			def.hasTimestamps = true
		}
	}

	return def
}
