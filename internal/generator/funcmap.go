package generator

import (
	"strings"
	"text/template"

	"github.com/forge-framework/forge/internal/parser"
)

// BuildFuncMap returns a FuncMap with all template helper functions.
func BuildFuncMap() template.FuncMap {
	return template.FuncMap{
		"goType":           goType,
		"goPointerType":    goPointerType,
		"lower":            lower,
		"plural":           plural,
		"camel":            camel,
		"snake":            snake,
		"hasModifier":      hasModifier,
		"getModifierValue": getModifierValue,
		"isRequired":       isRequired,
	}
}

// goType maps IR field type strings to Go type names.
func goType(fieldType string) string {
	switch fieldType {
	case "UUID":
		return "uuid.UUID"
	case "String":
		return "string"
	case "Text":
		return "string"
	case "Int":
		return "int"
	case "BigInt":
		return "int64"
	case "Decimal":
		return "decimal.Decimal"
	case "Bool":
		return "bool"
	case "DateTime":
		return "time.Time"
	case "Date":
		return "time.Time"
	case "Enum":
		return "string"
	case "JSON":
		return "json.RawMessage"
	case "Slug":
		return "string"
	case "Email":
		return "string"
	case "URL":
		return "string"
	default:
		return "interface{}"
	}
}

// goPointerType wraps goType in a pointer for optional/update fields.
func goPointerType(fieldType string) string {
	return "*" + goType(fieldType)
}

// lower converts a string to lowercase.
func lower(s string) string {
	return strings.ToLower(s)
}

// plural naively pluralizes a string.
func plural(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

// camel converts a string to PascalCase (first letter upper).
func camel(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// snake converts PascalCase to snake_case.
func snake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// hasModifier checks if a modifier with the given name exists in the list.
func hasModifier(modifiers []parser.ModifierIR, name string) bool {
	for _, m := range modifiers {
		if m.Type == name {
			return true
		}
	}
	return false
}

// getModifierValue retrieves the value of a modifier by name.
func getModifierValue(modifiers []parser.ModifierIR, name string) interface{} {
	for _, m := range modifiers {
		if m.Type == name {
			return m.Value
		}
	}
	return nil
}

// isRequired checks if the Required modifier is present.
func isRequired(modifiers []parser.ModifierIR) bool {
	return hasModifier(modifiers, "Required")
}
