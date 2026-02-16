// Package schema provides the DSL for defining resource schemas.
// This is the public API surface used by developers in resources/*/schema.go files.
//
// Schemas defined using this package are parseable by go/ast without requiring
// the gen/ directory to exist, supporting Forge's bootstrapping approach.
package schema

// SchemaItem is the common interface for all items that can be passed to Define().
// Fields, relationships, options, and timestamps all implement this interface.
type SchemaItem interface {
	// schemaItem is an unexported marker method to limit implementation
	// to types defined in this package.
	schemaItem()
}

// FieldType represents the type of a field in a schema.
type FieldType int

const (
	TypeUUID FieldType = iota
	TypeString
	TypeText
	TypeInt
	TypeBigInt
	TypeDecimal
	TypeBool
	TypeDateTime
	TypeDate
	TypeEnum
	TypeJSON
	TypeSlug
	TypeEmail
	TypeURL
)

// String returns the string representation of the field type.
func (ft FieldType) String() string {
	switch ft {
	case TypeUUID:
		return "UUID"
	case TypeString:
		return "String"
	case TypeText:
		return "Text"
	case TypeInt:
		return "Int"
	case TypeBigInt:
		return "BigInt"
	case TypeDecimal:
		return "Decimal"
	case TypeBool:
		return "Bool"
	case TypeDateTime:
		return "DateTime"
	case TypeDate:
		return "Date"
	case TypeEnum:
		return "Enum"
	case TypeJSON:
		return "JSON"
	case TypeSlug:
		return "Slug"
	case TypeEmail:
		return "Email"
	case TypeURL:
		return "URL"
	default:
		return "Unknown"
	}
}

// OnDeleteAction represents the action to take when a referenced record is deleted.
type OnDeleteAction int

const (
	Cascade OnDeleteAction = iota
	SetNull
	Restrict
	NoAction
)

// String returns the string representation of the on delete action.
func (a OnDeleteAction) String() string {
	switch a {
	case Cascade:
		return "CASCADE"
	case SetNull:
		return "SET NULL"
	case Restrict:
		return "RESTRICT"
	case NoAction:
		return "NO ACTION"
	default:
		return "NO ACTION"
	}
}
