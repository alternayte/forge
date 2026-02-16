---
phase: 01-foundation-schema-dsl
plan: 03
subsystem: parser
tags:
  - ast-parsing
  - go-parser
  - static-analysis
  - tdd
dependency_graph:
  requires:
    - 01-01 (schema DSL package)
    - 01-02 (IR types and error diagnostics)
  provides:
    - schema-ast-parser
    - literal-value-validation
    - batch-error-collection
  affects:
    - cli (plan 04 - will use parser to load schemas)
    - code-generation (phase 02 - will consume parsed IR)
tech_stack:
  added:
    - go/ast: AST traversal and inspection
    - go/parser: Go source file parsing
    - go/token: Position tracking for error reporting
  patterns:
    - TDD (test-driven development)
    - Visitor pattern (ast.Inspect for tree traversal)
    - Recursive descent (method chain extraction)
key_files:
  created:
    - internal/parser/parser.go: Entry points (ParseFile, ParseDir, ParseString)
    - internal/parser/extractor.go: AST extraction logic with chain handling
    - internal/parser/validator.go: Literal value validation
    - internal/parser/parser_test.go: Comprehensive TDD test suite (10 tests)
    - internal/parser/parser_integration_test.go: Product schema integration tests
  modified: []
decisions:
  - decision: Method chains traverse DOWN from outermost call
    rationale: "Arguments to Define() are outermost calls in chain (e.g., schema.String('Title').Required()); need to find root schema.X() call by recursing into Fun.X"
    impact: "findRootCall() helper extracts base constructor and field/relationship type"
  - decision: Collect all validation errors before returning
    rationale: "Developer fixes everything at once, no cascading noise (per user requirement)"
    impact: "Parser continues after errors, uses ast.Inspect to collect all issues in single pass"
  - decision: Separate extractField and extractRelationship functions
    rationale: "Different argument patterns (fields have 1 arg, relationships have 2) and different modifier handling"
    impact: "Cleaner code, easier to extend for new field/relationship types"
metrics:
  duration_minutes: 6.4
  completed_at: "2026-02-16T15:43:36Z"
  tasks_completed: 1
  files_created: 5
  tests_added: 12
  commits: 2
---

# Phase 01 Plan 03: Schema AST Parser Summary

**One-liner:** Go AST parser that reads schema.go files and extracts resource definitions into IR structs using static analysis, enabling "schema as source of truth" without requiring gen/ to exist

## Objective Achievement

Built the core technical foundation of Phase 1: a go/ast-based parser that statically analyzes schema.go files and extracts resource definitions into ResourceIR structs. This solves the bootstrapping constraint by parsing schemas before gen/ types exist.

**Must-haves delivered:**
- ✅ CLI can parse schema.go files using go/ast without requiring gen/ to exist
- ✅ Parser extracts schema.Define() calls, field definitions, modifiers, relationships, and options into IR
- ✅ Parser errors with clear message pointing to offending line when schema uses dynamic values
- ✅ Parser collects all errors in single pass — developer sees everything at once

## Tasks Completed

### Task 1: TDD RED - Write Failing Tests

**Commit:** `140a59a`

Created comprehensive test suite with 10 test cases covering all parser functionality:

1. TestParseSimpleResource - Basic resource with string fields and Required modifier
2. TestParseAllFieldTypes - All 14 field types (UUID, String, Text, Int, BigInt, Decimal, Bool, DateTime, Date, Enum, JSON, Slug, Email, URL)
3. TestParseFieldModifiersWithValues - Modifiers with arguments (MaxLen, MinLen, Default, Label, Placeholder)
4. TestParseRelationships - BelongsTo with Optional and OnDelete modifiers
5. TestParseResourceOptions - SoftDelete, Auditable, Timestamps
6. TestParseEnumWithValues - Enum field with allowed values and Default
7. TestRejectDynamicValues - Validates that variables produce rich diagnostics with error code E001
8. TestCollectMultipleErrors - Confirms multiple errors collected in single pass
9. TestParseFileWithNoSchemaDefine - Graceful handling of non-schema files
10. TestParseAllRelationshipTypes - BelongsTo, HasMany, HasOne, ManyToMany

Tests initially failed with "undefined: ParseString" as expected in RED phase.

### Task 2: TDD GREEN - Implement Parser

**Commit:** `c609d1f`

Implemented complete parser with three main components:

**parser.go (116 lines):**
- `ParseString(source, filename)` - Parse from string (for testing)
- `ParseFile(path)` - Parse single file from disk
- `ParseDir(dir)` - Find and parse all schema.go files in subdirectories
- Uses go/parser.ParseFile with ParseComments mode
- Reads source file content for rich error diagnostics

**extractor.go (444 lines):**
- `extractResources()` - Walk AST to find package-level var declarations with schema.Define() calls
- `extractSchemaDefinition()` - Extract resource name and categorize arguments into fields/relationships/options
- `findRootCall()` - Traverse method chains to find root schema.X() constructor
- `extractField()` - Extract FieldIR with type, name, modifiers, and enum values
- `extractRelationship()` - Extract RelationshipIR with type, name, table, and modifiers
- `extractModifiers()` - Recursively extract modifiers from method chains
- `extractLiteralValue()` - Extract values from BasicLit and SelectorExpr nodes

**Key implementation details:**
- Method chains represented as nested CallExpr nodes: `schema.String("Name").Required()` is `CallExpr{Fun: SelectorExpr{X: CallExpr{...}, Sel: Required}}`
- Must traverse DOWN from outermost call to find root constructor
- Uses strconv.Unquote() for string literals (BasicLit includes quotes)
- Handles schema.SetNull, schema.Cascade as SelectorExpr values
- Collects all diagnostics before returning (satisfies batch error requirement)

**validator.go (128 lines):**
- `validateLiteralValues()` - Check that all arguments are literals or schema.X selectors
- Uses ast.Inspect to recursively validate expressions
- Creates Diagnostic errors with line numbers, source context, and hints
- Rejects identifiers (variables) with error code E001
- Allows BasicLit (string/int/float), schema.X selectors, and schema.X() constructors

**Integration tests (parser_integration_test.go):**
- TestParseProductSchemaExample - Validates full Product schema from CONTEXT.md
- TestParseFileFromDisk - Tests real file I/O

All 12 tests pass (10 from TDD suite + 2 integration tests).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unused variables in extractModifiers**
- **Found during:** GREEN phase compilation
- **Issue:** Declared `modifiers` and `diagnostics` variables but returned call to extractModifiersFromChain immediately
- **Fix:** Removed unused variables, simplified function to direct return
- **Files modified:** internal/parser/extractor.go
- **Commit:** Included in c609d1f

No other deviations - plan executed exactly as written with TDD discipline.

## Technical Deep Dive

### AST Structure of Method Chains

Understanding method chain representation was critical:

```
schema.String("Title").Required().MaxLen(200)
```

AST structure:
```
CallExpr {
  Fun: SelectorExpr { X: CallExpr { ... }, Sel: "MaxLen" }
  Args: [BasicLit{Value: "200"}]
}
```

The outermost call is MaxLen, which wraps Required, which wraps String. To extract the field type, we must traverse DOWN to find the schema.String() call.

**findRootCall() algorithm:**
1. Start with outermost CallExpr (the argument to Define)
2. Check if Fun.X is an identifier (schema) - if yes, we found the root
3. If Fun.X is a CallExpr, recurse into it
4. Return the root call and the function name (field/relationship type)

### Error Collection Strategy

Per user requirement: "collect all errors in single pass - developer fixes everything at once."

Implementation:
- validateLiteralValues() uses ast.Inspect to walk entire expression tree
- Collects all Diagnostic errors into slice before returning
- extractSchemaDefinition() continues processing after validation errors
- ParseResult.Errors aggregates diagnostics from all resources

Result: Developer sees ALL schema errors in one report, no cascading fixes.

### Literal Validation Rules

Schema values must be statically analyzable. Allowed:
- BasicLit (string, int, float, bool literals)
- SelectorExpr pointing to schema package (schema.SetNull, schema.Cascade)
- CallExpr to schema constructors (nested fields)
- Identifiers for built-in constants (true, false, nil)

Rejected:
- Variable references (identifiers not in built-in list)
- Function calls outside schema package
- Arithmetic or logical expressions

Rejection produces Diagnostic with:
- Error code E001
- Line number and column
- Source line text
- Hint: "Replace dynamic value with a literal constant"

## Verification Results

**Build:** ✅ `go build ./internal/parser/` - zero errors

**Tests:** ✅ `go test ./internal/parser/` - all 12 tests passed

**Coverage:**
- ✅ All 14 field types parsed correctly
- ✅ All 4 relationship types parsed correctly
- ✅ All resource options extracted (SoftDelete, Auditable, TenantScoped, Searchable)
- ✅ Timestamps marker detected
- ✅ Field modifiers with values extracted (MaxLen, MinLen, Default, Label, Placeholder, Help)
- ✅ Relationship modifiers extracted (Optional, OnDelete)
- ✅ Enum values extracted from constructor arguments
- ✅ Dynamic values rejected with rich diagnostics
- ✅ Multiple errors collected in single pass
- ✅ Non-schema files gracefully skipped
- ✅ Product schema example from CONTEXT.md parses successfully

**Integration test results:**
```
Successfully parsed Product schema:
  - 6 fields
  - 2 relationships
  - SoftDelete: true
  - Timestamps: true
```

All fields, modifiers, relationships, and options extracted correctly from complex real-world schema.

## Example Usage

```go
// Parse a single file
result, err := parser.ParseFile("resources/product/schema.go")
if err != nil {
    log.Fatal(err)
}

if len(result.Errors) > 0 {
    for _, e := range result.Errors {
        fmt.Println(e) // Prints rich diagnostic with line numbers
    }
    os.Exit(1)
}

for _, resource := range result.Resources {
    fmt.Printf("Resource: %s\n", resource.Name)
    fmt.Printf("  Fields: %d\n", len(resource.Fields))
    fmt.Printf("  Relationships: %d\n", len(resource.Relationships))
}
```

## Success Criteria

All success criteria met:

✅ `ParseDir("resources/")` reads schema.go files from resource subdirectories
✅ Extracts all schema.Define() calls into ResourceIR structs
✅ Fields, modifiers, relationships, and options correctly populated
✅ Invalid schemas produce rich Diagnostic errors with file:line positions
✅ Fix suggestions included in diagnostics
✅ All errors collected in single pass

## Next Steps

**Immediate (Plan 04 - CLI skeleton):**
- Use parser to load schemas from resources/ directory
- Display parsed resources with UI styles
- Show diagnostics using error formatter

**Downstream Dependencies:**
- Phase 2 (code generation): Consume ResourceIR to generate models, routes, handlers
- Phase 3 (migrations): Use ResourceIR to generate migration DDL
- Phase 4 (admin UI): Use ResourceIR to generate CRUD interfaces

## Key Artifacts Delivered

| Artifact | Purpose | Interface |
|----------|---------|-----------|
| ParseString() | Parse schema from string (testing) | (source, filename) → (*ParseResult, error) |
| ParseFile() | Parse single schema file | (path) → (*ParseResult, error) |
| ParseDir() | Parse all schema.go in directory | (dir) → (*ParseResult, error) |
| findRootCall() | Extract base constructor from chain | (call) → (rootCall, funcName) |
| extractField() | Extract FieldIR from chained call | (fset, call, source, filename) → (*FieldIR, []Diagnostic) |
| extractRelationship() | Extract RelationshipIR | (fset, call, source, filename) → (*RelationshipIR, []Diagnostic) |
| validateLiteralValues() | Reject dynamic values | (fset, expr, source, filename) → []Diagnostic |

## Self-Check: PASSED

**Files verified:**
```
FOUND: internal/parser/parser.go
FOUND: internal/parser/extractor.go
FOUND: internal/parser/validator.go
FOUND: internal/parser/parser_test.go
FOUND: internal/parser/parser_integration_test.go
```

**Commits verified:**
```
FOUND: 140a59a (test: add failing tests for schema AST parser)
FOUND: c609d1f (feat: implement schema AST parser)
```

**Tests verified:**
```
PASS: TestParseProductSchemaExample
PASS: TestParseFileFromDisk
PASS: TestParseSimpleResource
PASS: TestParseAllFieldTypes
PASS: TestParseFieldModifiersWithValues
PASS: TestParseRelationships
PASS: TestParseResourceOptions
PASS: TestParseEnumWithValues
PASS: TestRejectDynamicValues
PASS: TestCollectMultipleErrors
PASS: TestParseFileWithNoSchemaDefine
PASS: TestParseAllRelationshipTypes
```

All claimed files, commits, and test results verified.
