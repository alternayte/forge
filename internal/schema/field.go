package schema

// Field represents a field in a schema definition.
type Field struct {
	name       string
	fieldType  FieldType
	modifiers  []Modifier
	enumValues []string // Used only for TypeEnum fields
}

// Name returns the field name.
func (f *Field) Name() string {
	return f.name
}

// Type returns the field type.
func (f *Field) Type() FieldType {
	return f.fieldType
}

// Modifiers returns all modifiers applied to this field.
func (f *Field) Modifiers() []Modifier {
	return f.modifiers
}

// EnumValues returns the allowed values for enum fields.
func (f *Field) EnumValues() []string {
	return f.enumValues
}

// schemaItem implements the SchemaItem interface.
func (f *Field) schemaItem() {}

// addModifier appends a modifier to the field's modifiers list.
func (f *Field) addModifier(mod Modifier) *Field {
	f.modifiers = append(f.modifiers, mod)
	return f
}

// Required marks the field as required (not nullable).
func (f *Field) Required() *Field {
	return f.addModifier(Modifier{Type: ModRequired})
}

// Optional marks the field as optional (nullable).
func (f *Field) Optional() *Field {
	return f.addModifier(Modifier{Type: ModOptional})
}

// PrimaryKey marks the field as the primary key.
func (f *Field) PrimaryKey() *Field {
	return f.addModifier(Modifier{Type: ModPrimaryKey})
}

// MaxLen sets the maximum length for string fields.
func (f *Field) MaxLen(n int) *Field {
	return f.addModifier(Modifier{Type: ModMaxLen, Value: n})
}

// MinLen sets the minimum length for string fields.
func (f *Field) MinLen(n int) *Field {
	return f.addModifier(Modifier{Type: ModMinLen, Value: n})
}

// Sortable marks the field as sortable in list views.
func (f *Field) Sortable() *Field {
	return f.addModifier(Modifier{Type: ModSortable})
}

// Filterable marks the field as filterable in list views.
func (f *Field) Filterable() *Field {
	return f.addModifier(Modifier{Type: ModFilterable})
}

// Searchable marks the field as searchable in full-text search.
func (f *Field) Searchable() *Field {
	return f.addModifier(Modifier{Type: ModSearchable})
}

// Unique marks the field as requiring unique values.
func (f *Field) Unique() *Field {
	return f.addModifier(Modifier{Type: ModUnique})
}

// Index marks the field for database indexing.
func (f *Field) Index() *Field {
	return f.addModifier(Modifier{Type: ModIndex})
}

// Default sets the default value for the field.
func (f *Field) Default(v interface{}) *Field {
	return f.addModifier(Modifier{Type: ModDefault, Value: v})
}

// Immutable marks the field as immutable after creation.
func (f *Field) Immutable() *Field {
	return f.addModifier(Modifier{Type: ModImmutable})
}

// Label sets the human-readable label for the field.
func (f *Field) Label(s string) *Field {
	return f.addModifier(Modifier{Type: ModLabel, Value: s})
}

// Placeholder sets the placeholder text for the field.
func (f *Field) Placeholder(s string) *Field {
	return f.addModifier(Modifier{Type: ModPlaceholder, Value: s})
}

// Help sets the help text for the field.
func (f *Field) Help(s string) *Field {
	return f.addModifier(Modifier{Type: ModHelp, Value: s})
}
