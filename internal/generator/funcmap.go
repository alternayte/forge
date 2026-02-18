package generator

import (
	"fmt"
	"strings"
	"text/template"
	"unicode"

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
		"zeroValue":              zeroValue,
		"enumValues":             enumValues,
		"hasMinLen":              hasMinLen,
		"getMinLen":              getMinLen,
		"getMaxLen":              getMaxLen,
		// API generation helpers
		"kebab":               kebab,
		"lowerCamel":          lowerCamel,
		"humaValidationTag":   humaValidationTag,
		"sortableFieldNames":  sortableFieldNames,
		"filterableFields":    filterableFields,
		"buildLinkHeader":     buildLinkHeader,
		"not":                 not,
		"join":                join,
		// HTML generation helpers
		"htmlInputType": htmlInputType,
		// Phase 7: Advanced data feature helpers
		"hasPermission":          hasPermission,
		"permissionRoles":        permissionRoles,
		"hasAnyVisibility":       hasAnyVisibility,
		"hasAnyPermission":       hasAnyPermission,
		"hasAuditableResource":   hasAuditableResource,
		"hasTenantScopedResource": hasTenantScopedResource,
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

// zeroValue returns the Go zero value for a field type.
func zeroValue(fieldType string) string {
	switch fieldType {
	case "UUID":
		return "uuid.UUID{}"
	case "String", "Text", "Enum", "Slug", "Email", "URL":
		return `""`
	case "Int":
		return "0"
	case "BigInt":
		return "0"
	case "Decimal":
		return "decimal.Zero"
	case "Bool":
		return "false"
	case "DateTime", "Date":
		return "time.Time{}"
	case "JSON":
		return "nil"
	default:
		return "nil"
	}
}

// enumValues returns the enum values for a field.
func enumValues(field parser.FieldIR) []string {
	return field.EnumValues
}

// hasMinLen checks if a field has a MinLen modifier.
func hasMinLen(modifiers []parser.ModifierIR) bool {
	return hasModifier(modifiers, "MinLen")
}

// getMinLen retrieves the MinLen modifier value.
func getMinLen(modifiers []parser.ModifierIR) interface{} {
	return getModifierValue(modifiers, "MinLen")
}

// getMaxLen retrieves the MaxLen modifier value.
func getMaxLen(modifiers []parser.ModifierIR) interface{} {
	return getModifierValue(modifiers, "MaxLen")
}

// kebab converts PascalCase to kebab-case (e.g., "BlogPost" -> "blog-post").
// Used for URL paths in route registration.
func kebab(s string) string {
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				result.WriteRune('-')
			} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				result.WriteRune('-')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}

// lowerCamel converts PascalCase to camelCase (e.g., "BlogPost" -> "blogPost").
// Used for operationIds in OpenAPI spec generation.
func lowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	return strings.ToLower(string(runes[0:1])) + string(runes[1:])
}

// humaValidationTag builds Huma validation tag string from field modifiers.
// Maps MinLen -> minLength, MaxLen -> maxLength for Huma struct tags.
func humaValidationTag(field parser.FieldIR) string {
	var parts []string

	if minLen := getModifierValue(field.Modifiers, "MinLen"); minLen != nil {
		parts = append(parts, fmt.Sprintf(`minLength:"%v"`, minLen))
	}
	if maxLen := getModifierValue(field.Modifiers, "MaxLen"); maxLen != nil {
		parts = append(parts, fmt.Sprintf(`maxLength:"%v"`, maxLen))
	}
	if min := getModifierValue(field.Modifiers, "Min"); min != nil {
		parts = append(parts, fmt.Sprintf(`minimum:"%v"`, min))
	}
	if max := getModifierValue(field.Modifiers, "Max"); max != nil {
		parts = append(parts, fmt.Sprintf(`maximum:"%v"`, max))
	}
	if field.Type == "Enum" && len(field.EnumValues) > 0 {
		parts = append(parts, fmt.Sprintf(`enum:"%s"`, strings.Join(field.EnumValues, ",")))
	}

	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// sortableFieldNames returns comma-separated snake_case names of sortable fields.
// Used for the enum tag on the sort query parameter.
func sortableFieldNames(fields []parser.FieldIR) string {
	var names []string
	for _, f := range fields {
		if isSortable(f.Modifiers) {
			names = append(names, snake(f.Name))
		}
	}
	return strings.Join(names, ",")
}

// filterableFields returns only fields with the Filterable modifier.
func filterableFields(fields []parser.FieldIR) []parser.FieldIR {
	var result []parser.FieldIR
	for _, f := range fields {
		if isFilterable(f.Modifiers) {
			result = append(result, f)
		}
	}
	return result
}

// buildLinkHeader builds an RFC 8288 Link header value for cursor pagination.
// Returns the header string: <{basePath}?cursor={cursor}&limit={limit}>; rel="next"
func buildLinkHeader(basePath string, cursor string, limit int) string {
	return fmt.Sprintf(`<%s?cursor=%s&limit=%d>; rel="next"`, basePath, cursor, limit)
}

// not returns the boolean negation of its argument.
func not(b bool) bool {
	return !b
}

// join joins a string slice with a separator.
func join(sep string, s []string) string {
	return strings.Join(s, sep)
}

// Phase 7: Advanced data feature helpers

// hasPermission returns true if the resource options define the given operation.
func hasPermission(opts parser.ResourceOptionsIR, operation string) bool {
	if opts.Permissions == nil {
		return false
	}
	_, ok := opts.Permissions[operation]
	return ok
}

// permissionRoles returns a Go code literal of quoted role strings for the
// given operation (e.g., `"admin", "editor"`). Returns an empty string if
// the operation has no permissions defined.
func permissionRoles(opts parser.ResourceOptionsIR, operation string) string {
	if opts.Permissions == nil {
		return ""
	}
	roles, ok := opts.Permissions[operation]
	if !ok || len(roles) == 0 {
		return ""
	}
	quoted := make([]string, len(roles))
	for i, r := range roles {
		quoted[i] = `"` + r + `"`
	}
	return strings.Join(quoted, ", ")
}

// hasAnyVisibility returns true if any field in the list has a Visibility modifier.
func hasAnyVisibility(fields []parser.FieldIR) bool {
	for _, f := range fields {
		if hasModifier(f.Modifiers, "Visibility") {
			return true
		}
	}
	return false
}

// hasAnyPermission returns true if the resource options contain any permission rules.
func hasAnyPermission(opts parser.ResourceOptionsIR) bool {
	return len(opts.Permissions) > 0
}

// hasAuditableResource returns true if any resource in the slice has Auditable enabled.
func hasAuditableResource(resources []parser.ResourceIR) bool {
	for _, r := range resources {
		if r.Options.Auditable {
			return true
		}
	}
	return false
}

// hasTenantScopedResource returns true if any resource in the slice has TenantScoped enabled.
func hasTenantScopedResource(resources []parser.ResourceIR) bool {
	for _, r := range resources {
		if r.Options.TenantScoped {
			return true
		}
	}
	return false
}

// htmlInputType maps IR field type strings to HTML input type attributes.
// Used in HTML form generation to select the appropriate input element type.
func htmlInputType(fieldType string) string {
	switch fieldType {
	case "String":
		return "text"
	case "Text":
		return "text"
	case "Int":
		return "number"
	case "BigInt":
		return "number"
	case "Decimal":
		return "number"
	case "Bool":
		return "checkbox"
	case "Email":
		return "email"
	case "URL":
		return "url"
	case "Date":
		return "date"
	case "DateTime":
		return "datetime-local"
	default:
		return "text"
	}
}
