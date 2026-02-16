package schema

// ModifierType represents the type of modifier applied to a field.
type ModifierType int

const (
	ModRequired ModifierType = iota
	ModOptional
	ModMaxLen
	ModMinLen
	ModSortable
	ModFilterable
	ModSearchable
	ModUnique
	ModIndex
	ModDefault
	ModImmutable
	ModLabel
	ModPlaceholder
	ModHelp
	ModPrimaryKey
)

// Modifier represents a modification or constraint applied to a field.
type Modifier struct {
	Type  ModifierType
	Value interface{} // Used for MaxLen, MinLen, Default, Label, Placeholder, Help
}
