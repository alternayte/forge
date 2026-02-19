package parser

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/forge-framework/forge/internal/errors"
)

// extractResources walks the AST and extracts all schema.Define() calls.
func extractResources(fset *token.FileSet, file *ast.File, source []byte, filename string) ([]ResourceIR, []errors.Diagnostic) {
	var resources []ResourceIR
	var diagnostics []errors.Diagnostic

	// Walk the AST looking for package-level var declarations
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.VAR {
			continue
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valueSpec.Values) == 0 {
				continue
			}

			// Check if the value is a schema.Define() call
			callExpr, ok := valueSpec.Values[0].(*ast.CallExpr)
			if !ok {
				continue
			}

			// Check if this is schema.Define
			if !isSchemaDefineCall(callExpr) {
				continue
			}

			// Extract the resource definition
			resource, diags := extractSchemaDefinition(fset, callExpr, source, filename)
			diagnostics = append(diagnostics, diags...)

			if resource != nil {
				resources = append(resources, *resource)
			}
		}
	}

	return resources, diagnostics
}

// isSchemaDefineCall checks if a call expression is schema.Define()
func isSchemaDefineCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "schema" && sel.Sel.Name == "Define"
}

// findRootCall traverses a method chain to find the root schema.X() call.
// Returns the root call and the function name.
func findRootCall(call *ast.CallExpr) (*ast.CallExpr, string) {
	current := call

	// Walk down the chain until we find schema.X()
	for {
		sel, ok := current.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil, ""
		}

		// Check if X is an identifier (schema)
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "schema" {
			return current, sel.Sel.Name
		}

		// Check if X is another call expression (chained method)
		innerCall, ok := sel.X.(*ast.CallExpr)
		if !ok {
			return nil, ""
		}

		current = innerCall
	}
}

// extractSchemaDefinition extracts a ResourceIR from a schema.Define() call.
func extractSchemaDefinition(fset *token.FileSet, call *ast.CallExpr, source []byte, filename string) (*ResourceIR, []errors.Diagnostic) {
	var diagnostics []errors.Diagnostic

	if len(call.Args) == 0 {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"schema.Define() requires at least a resource name",
		).File(filename).Line(fset.Position(call.Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	// First argument must be a string literal (resource name)
	nameLit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || nameLit.Kind != token.STRING {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"schema.Define() first argument must be a string literal resource name",
		).File(filename).Line(fset.Position(call.Args[0].Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	// Unquote the string literal
	name, err := strconv.Unquote(nameLit.Value)
	if err != nil {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"failed to parse resource name",
		).File(filename).Line(fset.Position(nameLit.Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	resource := &ResourceIR{
		Name:       name,
		SourceFile: filename,
		SourceLine: fset.Position(call.Pos()).Line,
	}

	// Process remaining arguments (fields, relationships, options)
	for i := 1; i < len(call.Args); i++ {
		arg := call.Args[i]

		// Validate that arguments are literals
		valDiags := validateLiteralValues(fset, arg, source, filename)
		diagnostics = append(diagnostics, valDiags...)

		// Extract based on type
		argCall, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}

		// Find the root call in the chain (the schema.X() call)
		rootCall, funcName := findRootCall(argCall)
		if funcName == "" {
			continue
		}

		// Check if it's a field type
		if isFieldType(funcName) {
			field, fieldDiags := extractField(fset, argCall, source, filename)
			diagnostics = append(diagnostics, fieldDiags...)
			if field != nil {
				resource.Fields = append(resource.Fields, *field)
			}
		} else if isRelationshipType(funcName) {
			rel, relDiags := extractRelationship(fset, argCall, source, filename)
			diagnostics = append(diagnostics, relDiags...)
			if rel != nil {
				resource.Relationships = append(resource.Relationships, *rel)
			}
		} else if isOptionType(funcName) {
			extractOption(funcName, &resource.Options)
		} else if funcName == "Timestamps" {
			resource.HasTimestamps = true
		} else if isPermissionType(funcName) {
			op, roles := extractPermission(fset, argCall, source, filename)
			if op != "" && len(roles) > 0 {
				if resource.Options.Permissions == nil {
					resource.Options.Permissions = make(PermissionsIR)
				}
				resource.Options.Permissions[op] = roles
			}
		} else if isHooksType(funcName) {
			hooks, err := extractHooks(argCall)
			if err == nil {
				resource.Options.Hooks = hooks
			}
		}

		// Use rootCall if needed for future enhancements
		_ = rootCall
	}

	return resource, diagnostics
}

// isFieldType checks if a function name is a field type constructor.
func isFieldType(name string) bool {
	fieldTypes := map[string]bool{
		"UUID": true, "String": true, "Text": true, "Int": true,
		"BigInt": true, "Decimal": true, "Bool": true, "DateTime": true,
		"Date": true, "Enum": true, "JSON": true, "Slug": true,
		"Email": true, "URL": true,
	}
	return fieldTypes[name]
}

// isRelationshipType checks if a function name is a relationship type.
func isRelationshipType(name string) bool {
	return name == "BelongsTo" || name == "HasMany" || name == "HasOne" || name == "ManyToMany"
}

// isOptionType checks if a function name is an option type.
func isOptionType(name string) bool {
	return name == "SoftDelete" || name == "Auditable" || name == "TenantScoped" || name == "Searchable"
}

// extractOption sets the appropriate option flag.
func extractOption(name string, options *ResourceOptionsIR) {
	switch name {
	case "SoftDelete":
		options.SoftDelete = true
	case "Auditable":
		options.Auditable = true
	case "TenantScoped":
		options.TenantScoped = true
	case "Searchable":
		options.Searchable = true
	}
}

// extractField extracts a FieldIR from a field constructor call.
// The call may be a chained expression like schema.String("Name").Required()
func extractField(fset *token.FileSet, call *ast.CallExpr, source []byte, filename string) (*FieldIR, []errors.Diagnostic) {
	var diagnostics []errors.Diagnostic

	// Find the root schema.X() call in the chain
	rootCall, fieldType := findRootCall(call)
	if rootCall == nil || fieldType == "" {
		return nil, diagnostics
	}

	// First argument is the field name
	if len(rootCall.Args) == 0 {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"field constructor requires a name argument",
		).File(filename).Line(fset.Position(rootCall.Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	// Extract field name
	nameLit, ok := rootCall.Args[0].(*ast.BasicLit)
	if !ok || nameLit.Kind != token.STRING {
		diag := errors.NewDiagnostic(
			errors.ErrInvalidFieldName,
			"field name must be a string literal",
		).File(filename).Line(fset.Position(rootCall.Args[0].Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	name, err := strconv.Unquote(nameLit.Value)
	if err != nil {
		diag := errors.NewDiagnostic(
			errors.ErrInvalidFieldName,
			"failed to parse field name",
		).File(filename).Line(fset.Position(nameLit.Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	field := &FieldIR{
		Name:       name,
		Type:       fieldType,
		SourceLine: fset.Position(rootCall.Pos()).Line,
	}

	// For Enum fields, extract the enum values (from rootCall args after the name)
	if fieldType == "Enum" {
		for i := 1; i < len(rootCall.Args); i++ {
			valLit, ok := rootCall.Args[i].(*ast.BasicLit)
			if ok && valLit.Kind == token.STRING {
				val, err := strconv.Unquote(valLit.Value)
				if err == nil {
					field.EnumValues = append(field.EnumValues, val)
				}
			}
		}
	}

	// Extract modifiers from method chain
	modifiers, modDiags := extractModifiers(fset, call, source, filename)
	diagnostics = append(diagnostics, modDiags...)
	field.Modifiers = modifiers

	return field, diagnostics
}

// extractRelationship extracts a RelationshipIR from a relationship constructor call.
// The call may be chained like schema.BelongsTo("User", "users").Optional()
func extractRelationship(fset *token.FileSet, call *ast.CallExpr, source []byte, filename string) (*RelationshipIR, []errors.Diagnostic) {
	var diagnostics []errors.Diagnostic

	// Find the root schema.X() call in the chain
	rootCall, relType := findRootCall(call)
	if rootCall == nil || relType == "" {
		return nil, diagnostics
	}

	// First argument is the relationship name, second is table
	if len(rootCall.Args) < 2 {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"relationship constructor requires name and table arguments",
		).File(filename).Line(fset.Position(rootCall.Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	// Extract name
	nameLit, ok := rootCall.Args[0].(*ast.BasicLit)
	if !ok || nameLit.Kind != token.STRING {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"relationship name must be a string literal",
		).File(filename).Line(fset.Position(rootCall.Args[0].Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	name, _ := strconv.Unquote(nameLit.Value)

	// Extract table
	tableLit, ok := rootCall.Args[1].(*ast.BasicLit)
	if !ok || tableLit.Kind != token.STRING {
		diag := errors.NewDiagnostic(
			errors.ErrMissingArgument,
			"relationship table must be a string literal",
		).File(filename).Line(fset.Position(rootCall.Args[1].Pos()).Line).Build()
		return nil, []errors.Diagnostic{diag}
	}

	table, _ := strconv.Unquote(tableLit.Value)

	rel := &RelationshipIR{
		Name:       name,
		Type:       relType,
		Table:      table,
		SourceLine: fset.Position(rootCall.Pos()).Line,
	}

	// Extract modifiers (Optional, OnDelete)
	modifiers, modDiags := extractModifiers(fset, call, source, filename)
	diagnostics = append(diagnostics, modDiags...)

	for _, mod := range modifiers {
		if mod.Type == "Optional" {
			rel.Optional = true
		} else if mod.Type == "OnDelete" {
			if val, ok := mod.Value.(string); ok {
				rel.OnDelete = val
			}
		} else if mod.Type == "Eager" {
			rel.Eager = true
		}
	}

	return rel, diagnostics
}

// extractModifiers extracts modifiers from a method chain.
// Method chains in AST are represented as nested expressions.
func extractModifiers(fset *token.FileSet, rootCall *ast.CallExpr, source []byte, filename string) ([]ModifierIR, []errors.Diagnostic) {
	// The call we're given is the outermost call in the chain (the argument to Define).
	// We walk DOWN from rootCall to collect modifiers.
	return extractModifiersFromChain(fset, rootCall, source, filename)
}

// extractModifiersFromChain recursively extracts modifiers from a chained call.
func extractModifiersFromChain(fset *token.FileSet, call *ast.CallExpr, source []byte, filename string) ([]ModifierIR, []errors.Diagnostic) {
	var modifiers []ModifierIR
	var diagnostics []errors.Diagnostic

	// Check if this call's function is a selector expression (method call)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return modifiers, diagnostics
	}

	// If the selector's X is a call expression, it's a chained method
	innerCall, isChained := sel.X.(*ast.CallExpr)
	if isChained {
		// Recursively extract modifiers from the inner call
		innerMods, innerDiags := extractModifiersFromChain(fset, innerCall, source, filename)
		modifiers = append(modifiers, innerMods...)
		diagnostics = append(diagnostics, innerDiags...)
	}

	// Now extract the modifier from this level
	methodName := sel.Sel.Name

	// Check if this is a modifier method
	if isModifierMethod(methodName) {
		modifier := ModifierIR{
			Type:       methodName,
			SourceLine: fset.Position(call.Pos()).Line,
		}

		// Extract argument if present
		if len(call.Args) > 0 {
			value, diags := extractLiteralValue(fset, call.Args[0], source, filename)
			modifier.Value = value
			diagnostics = append(diagnostics, diags...)
		}

		modifiers = append(modifiers, modifier)
	}

	return modifiers, diagnostics
}

// isModifierMethod checks if a method name is a modifier.
func isModifierMethod(name string) bool {
	modifiers := map[string]bool{
		"Required": true, "Optional": true, "PrimaryKey": true,
		"MaxLen": true, "MinLen": true, "Sortable": true,
		"Filterable": true, "Searchable": true, "Unique": true,
		"Index": true, "Default": true, "Immutable": true,
		"Label": true, "Placeholder": true, "Help": true,
		"OnDelete": true,
		// Phase 7: advanced data feature modifiers
		"Visibility": true, "Mutability": true, "Eager": true,
	}
	return modifiers[name]
}

// isPermissionType checks if a function name is a Permission constructor.
func isPermissionType(name string) bool {
	return name == "Permission"
}

// isHooksType checks if a function name is a WithHooks constructor.
func isHooksType(name string) bool {
	return name == "WithHooks"
}

// extractHooks extracts a HooksIR from a schema.WithHooks() call expression.
//
// AST structure of schema.WithHooks(schema.Hooks{AfterCreate: []schema.JobRef{{Kind: "x", Queue: "y"}}}):
//   - CallExpr with Fun = SelectorExpr(schema.WithHooks)
//   - Args[0] = CompositeLit (type schema.Hooks)
//     - Elts: KeyValueExpr (Key="AfterCreate"|"AfterUpdate", Value=CompositeLit of []schema.JobRef)
//       - Elts: CompositeLit (type schema.JobRef) with KeyValueExpr for Kind and Queue
func extractHooks(callExpr *ast.CallExpr) (HooksIR, error) {
	var hooks HooksIR

	if len(callExpr.Args) == 0 {
		return hooks, nil
	}

	// Args[0] should be a composite literal of type schema.Hooks
	hooksLit, ok := callExpr.Args[0].(*ast.CompositeLit)
	if !ok {
		return hooks, nil
	}

	// Iterate over fields of the schema.Hooks composite literal
	for _, elt := range hooksLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		// Key must be an identifier (AfterCreate or AfterUpdate)
		keyIdent, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		fieldName := keyIdent.Name

		// Value is a composite literal of []schema.JobRef
		jobRefSlice, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}

		var jobRefs []JobRefIR
		for _, jobElt := range jobRefSlice.Elts {
			// Each element is a composite literal for a schema.JobRef struct
			jobLit, ok := jobElt.(*ast.CompositeLit)
			if !ok {
				continue
			}

			var ref JobRefIR
			for _, jobField := range jobLit.Elts {
				jobKV, ok := jobField.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				fieldKey, ok := jobKV.Key.(*ast.Ident)
				if !ok {
					continue
				}
				valLit, ok := jobKV.Value.(*ast.BasicLit)
				if !ok || valLit.Kind != token.STRING {
					continue
				}
				val, err := strconv.Unquote(valLit.Value)
				if err != nil {
					continue
				}
				switch fieldKey.Name {
				case "Kind":
					ref.Kind = val
				case "Queue":
					ref.Queue = val
				}
			}
			jobRefs = append(jobRefs, ref)
		}

		switch fieldName {
		case "AfterCreate":
			hooks.AfterCreate = jobRefs
		case "AfterUpdate":
			hooks.AfterUpdate = jobRefs
		}
	}

	return hooks, nil
}

// extractPermission extracts the operation and roles from a schema.Permission() call.
func extractPermission(fset *token.FileSet, call *ast.CallExpr, source []byte, filename string) (string, []string) {
	rootCall, _ := findRootCall(call)
	if rootCall == nil || len(rootCall.Args) < 2 {
		return "", nil
	}
	// First arg: operation string
	opLit, ok := rootCall.Args[0].(*ast.BasicLit)
	if !ok || opLit.Kind != token.STRING {
		return "", nil
	}
	op, _ := strconv.Unquote(opLit.Value)
	// Remaining args: role strings
	var roles []string
	for i := 1; i < len(rootCall.Args); i++ {
		roleLit, ok := rootCall.Args[i].(*ast.BasicLit)
		if ok && roleLit.Kind == token.STRING {
			role, _ := strconv.Unquote(roleLit.Value)
			roles = append(roles, role)
		}
	}
	return op, roles
}

// extractLiteralValue extracts a value from a literal expression.
func extractLiteralValue(fset *token.FileSet, expr ast.Expr, source []byte, filename string) (interface{}, []errors.Diagnostic) {
	var diagnostics []errors.Diagnostic

	switch lit := expr.(type) {
	case *ast.BasicLit:
		switch lit.Kind {
		case token.STRING:
			val, err := strconv.Unquote(lit.Value)
			if err != nil {
				return nil, diagnostics
			}
			return val, diagnostics
		case token.INT:
			val, err := strconv.Atoi(lit.Value)
			if err != nil {
				return nil, diagnostics
			}
			return val, diagnostics
		case token.FLOAT:
			val, err := strconv.ParseFloat(lit.Value, 64)
			if err != nil {
				return nil, diagnostics
			}
			return val, diagnostics
		}
	case *ast.SelectorExpr:
		// Handle schema.SetNull, schema.Cascade, etc.
		if ident, ok := lit.X.(*ast.Ident); ok && ident.Name == "schema" {
			return lit.Sel.Name, diagnostics
		}
	}

	return nil, diagnostics
}
