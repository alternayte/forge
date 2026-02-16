# Phase 1: Foundation & Schema DSL - Context

**Gathered:** 2026-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Bootstrapping architecture and schema definition API. Developer can define resource schemas with field types, modifiers, and relationships that are parseable via go/ast without gen/ existing. CLI can initialize projects, parse schemas, and manage tool binaries. Code generation and migration execution are separate phases.

</domain>

<decisions>
## Implementation Decisions

### Schema authoring feel
- Fluent method chaining for field definitions: `schema.String("Title").Required().MaxLen(200)`
- Package-level variables for resource definitions: `var Post = schema.Define("Post", ...)`
- `schema.Define()` takes variadic args — fields, relationships, timestamps all at same level in flat list
- Compact style: one field per line, prioritize scannability
- Relationships are inline entries alongside fields, not a separate block: `schema.BelongsTo("Category", "categories").Optional().OnDelete(schema.SetNull)`

### CLI behavior & output
- `forge init` creates a full starter project: forge.toml, resources/ directory, example schema, main.go, go.mod, .gitignore, README — ready to `forge generate` immediately
- `forge init my-project` creates new directory; `forge init` (no arg) initializes current directory using its name
- Polished styled output: colors, icons (checkmarks/crosses), grouped by category — similar feel to `cargo build` or `next dev`
- Tool sync is on-demand: binaries (templ, sqlc, tailwind, atlas) downloaded only when a command needs them, not upfront
- `forge generate` works offline — no database connection required. Only `forge migrate` and `forge db` commands need a live DB

### Error experience
- Rust-style rich errors: show the offending line, underline the problem, suggest a fix — file:line, code snippet, "did you mean...?"
- Collect all errors in a single parse pass — developer fixes everything at once, no cascading noise
- Dynamic value errors explain the constraint: "Forge schemas must use literal values for static analysis. Found variable 'maxLen' — use a constant or literal instead." Teach the why, not just the what
- Each error type includes a reference link or error code (e.g., "See: forge.dev/errors/E001" or `forge help schema-errors`) for deeper context

### Project structure
- Resource-colocated layout: each resource is its own package directory under `resources/` — schema.go, handlers.go, form.templ all live together per resource
- Generated code in structured `gen/` subdirectories: gen/models/, gen/queries/, gen/atlas/, gen/handlers/ — subdirectories by concern
- forge.toml uses commented template style: all available options present but commented out with defaults shown — developer uncomments to customize

### Claude's Discretion
- Exact field type set and modifier names (String, Int, UUID, etc.)
- Internal IR (intermediate representation) structure from go/ast parsing
- How tool sync detects platform and downloads binaries
- Error code numbering scheme
- Exact forge.toml section organization

</decisions>

<specifics>
## Specific Ideas

- Schema syntax inspired by PRD example: `schema.Define("Product", schema.Options{...}, schema.UUID("ID").PrimaryKey(), schema.String("Title").Required().MaxLen(200), schema.BelongsTo("Category", "categories"), schema.Timestamps())`
- CLI output should feel like Cargo — professional, not noisy
- Error messages should teach developers about the static analysis constraint, not just reject their code
- "I want to be able to find photos by roughly when they were taken" ethos applied to resources: `resources/product/schema.go` — you find the schema where the resource lives

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 01-foundation-schema-dsl*
*Context gathered: 2026-02-16*
