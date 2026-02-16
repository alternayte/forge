package generator

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/forge-framework/forge/internal/parser"
)

// BuildFuncMap returns a FuncMap with all template helper functions.
func BuildFuncMap() template.FuncMap {
	return template.FuncMap{
		"goType":                 goType,
		"goPointerType":          goPointerType,
		"lower":                  lower,
		"plural":                 plural,
		"camel":                  camel,
		"snake":                  snake,
		"hasModifier":            hasModifier,
		"getModifierValue":       getModifierValue,
		"isRequired":             isRequired,
		"isIDField":              isIDField,
		"isFilterable":           isFilterable,
		"isSortable":             isSortable,
		"atlasType":              atlasType,
		"atlasTypeWithModifiers": atlasTypeWithModifiers,
		"atlasNull":              atlasNull,
		"atlasDefault":           atlasDefault,
		"hasDefault":             hasDefault,
		"defaultTestValue":       defaultTestValue,
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
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		// Add underscore before uppercase letter if:
		// 1. Not the first character
		// 2. Previous character is lowercase OR next character is lowercase (handles HTTPStatus -> http_status)
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			// Add underscore if previous was lowercase or if this is start of acronym followed by lowercase
			if prev >= 'a' && prev <= 'z' {
				result.WriteRune('_')
			} else if i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
				result.WriteRune('_')
			}
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

// isIDField checks if a field is the ID field (name="ID" and type="UUID").
func isIDField(field parser.FieldIR) bool {
	return field.Name == "ID" && field.Type == "UUID"
}

// isFilterable checks if a field has the Filterable modifier.
func isFilterable(modifiers []parser.ModifierIR) bool {
	return hasModifier(modifiers, "Filterable")
}

// isSortable checks if a field has the Sortable modifier.
func isSortable(modifiers []parser.ModifierIR) bool {
	return hasModifier(modifiers, "Sortable")
}

// atlasType maps IR field type strings to Atlas HCL/PostgreSQL type names.
func atlasType(fieldType string) string {
	switch fieldType {
	case "UUID":
		return "uuid"
	case "String":
		return "varchar(255)"
	case "Text":
		return "text"
	case "Int":
		return "integer"
	case "BigInt":
		return "bigint"
	case "Decimal":
		return "numeric(10,2)"
	case "Bool":
		return "boolean"
	case "DateTime":
		return "timestamptz"
	case "Date":
		return "date"
	case "Enum":
		return "text"
	case "JSON":
		return "jsonb"
	case "Slug":
		return "varchar(255)"
	case "Email":
		return "varchar(255)"
	case "URL":
		return "text"
	default:
		return "text"
	}
}

// atlasTypeWithModifiers returns the Atlas type with modifiers applied (e.g., MaxLen).
func atlasTypeWithModifiers(field parser.FieldIR) string {
	baseType := atlasType(field.Type)

	// Check for MaxLen modifier to override varchar length
	if field.Type == "String" || field.Type == "Slug" || field.Type == "Email" {
		if maxLen := getModifierValue(field.Modifiers, "MaxLen"); maxLen != nil {
			if length, ok := maxLen.(int); ok {
				return fmt.Sprintf("varchar(%d)", length)
			}
		}
	}

	return baseType
}

// atlasNull returns "false" if Required modifier present, "true" otherwise.
func atlasNull(modifiers []parser.ModifierIR) string {
	if hasModifier(modifiers, "Required") {
		return "false"
	}
	return "true"
}

// atlasDefault returns the Atlas default expression if Default modifier is present.
func atlasDefault(field parser.FieldIR) string {
	defaultVal := getModifierValue(field.Modifiers, "Default")
	if defaultVal == nil {
		return ""
	}

	switch v := defaultVal.(type) {
	case string:
		// String defaults are quoted
		return `"` + v + `"`
	case bool:
		// Bool defaults are true/false
		if v {
			return "true"
		}
		return "false"
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// hasDefault checks if a field has a Default modifier.
func hasDefault(field parser.FieldIR) bool {
	return hasModifier(field.Modifiers, "Default")
}

// defaultTestValue returns a reasonable default test value for a field type.
func defaultTestValue(field parser.FieldIR) string {
	switch field.Type {
	case "UUID":
		return "uuid.New()"
	case "String":
		return fmt.Sprintf(`"test-%s"`, snake(field.Name))
	case "Text":
		return fmt.Sprintf(`"Test %s content"`, field.Name)
	case "Int":
		return "42"
	case "BigInt":
		return "100000"
	case "Decimal":
		return "decimal.NewFromFloat(9.99)"
	case "Bool":
		return "true"
	case "DateTime":
		return "time.Now()"
	case "Date":
		return "time.Now()"
	case "Enum":
		// Use first enum value if available, otherwise "default"
		if len(field.EnumValues) > 0 {
			return fmt.Sprintf(`"%s"`, field.EnumValues[0])
		}
		return `"default"`
	case "JSON":
		return `json.RawMessage("{}")`
	case "Slug":
		return `"test-slug"`
	case "Email":
		return `"test@example.com"`
	case "URL":
		return `"https://example.com"`
	default:
		return `""`
	}
}
