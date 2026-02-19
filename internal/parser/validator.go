package parser

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/alternayte/forge/internal/errors"
)

// validateLiteralValues checks that all arguments are literal values or schema.X selectors.
// Returns diagnostics for any dynamic values (variables, function calls, etc.).
func validateLiteralValues(fset *token.FileSet, expr ast.Expr, source []byte, filename string) []errors.Diagnostic {
	var diagnostics []errors.Diagnostic

	// Recursively validate the expression and its children
	ast.Inspect(expr, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		// Check for identifier references (variables)
		if ident, ok := n.(*ast.Ident); ok {
			// Skip built-in identifiers and type names
			if isBuiltinOrType(ident.Name) {
				return true
			}

			// Check if this identifier is used as an argument value (not as part of a selector)
			// We need to check the parent context, but since we don't have it in ast.Inspect,
			// we'll be conservative and flag identifiers that look like variables

			// Skip if it's part of a selector expression (like schema.SetNull)
			// We can't easily check this in ast.Inspect without parent tracking,
			// so we'll check in the call expression validation instead
			return true
		}

		// Check call expressions for argument validation
		if call, ok := n.(*ast.CallExpr); ok {
			for _, arg := range call.Args {
				if !isLiteralOrSchemaSelector(arg) {
					// This is a dynamic value - create diagnostic
					pos := fset.Position(arg.Pos())
					sourceLine := extractLineFromSourceBytes(source, pos.Line)

					// Calculate underline position
					col := pos.Column - 1 // 0-based
					underlineLen := len(extractExprText(arg, source))

					diag := errors.NewDiagnostic(
						errors.ErrDynamicValue,
						"schema values must be literal constants, not variables or expressions",
					).
						File(filename).
						Line(pos.Line).
						Column(pos.Column).
						SourceLine(sourceLine).
						Underline(col, underlineLen).
						Hint("Replace dynamic value with a literal constant (string, number, or schema.Constant)").
						Build()

					diagnostics = append(diagnostics, diag)
				}
			}
		}

		return true
	})

	return diagnostics
}

// isLiteralOrSchemaSelector checks if an expression is a literal value or schema.X selector.
func isLiteralOrSchemaSelector(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.BasicLit:
		// String, int, float, bool literals are allowed
		return true
	case *ast.SelectorExpr:
		// schema.SetNull, schema.Cascade, etc. are allowed
		if ident, ok := e.X.(*ast.Ident); ok {
			return ident.Name == "schema"
		}
		return false
	case *ast.CallExpr:
		// Nested schema constructor calls are allowed
		if sel, ok := e.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name == "schema"
			}
		}
		return false
	case *ast.Ident:
		// Only built-in constants like true, false, nil
		return e.Name == "true" || e.Name == "false" || e.Name == "nil"
	default:
		return false
	}
}

// isBuiltinOrType checks if an identifier is a built-in or type name.
func isBuiltinOrType(name string) bool {
	builtins := map[string]bool{
		"true": true, "false": true, "nil": true,
		"bool": true, "string": true, "int": true, "float64": true,
	}
	return builtins[name]
}

// extractLineFromSourceBytes extracts a specific line from source bytes.
func extractLineFromSourceBytes(source []byte, lineNum int) string {
	lines := strings.Split(string(source), "\n")
	if lineNum > 0 && lineNum <= len(lines) {
		return lines[lineNum-1]
	}
	return ""
}

// extractExprText attempts to extract the text representation of an expression.
func extractExprText(expr ast.Expr, source []byte) string {
	// Simple heuristic: if it's an identifier, return its name
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	// For other expressions, return a placeholder
	return "expression"
}
