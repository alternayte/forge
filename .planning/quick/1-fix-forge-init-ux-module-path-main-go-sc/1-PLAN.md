---
phase: quick-1
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/scaffold/scaffold.go
  - internal/scaffold/templates/main.go.tmpl
  - internal/cli/dev.go
autonomous: true
requirements: [UX-01, UX-02, UX-03]

must_haves:
  truths:
    - "forge init myapp produces go.mod with module myapp (not github.com/someone/myapp)"
    - "forge init in existing directory uses directory name as module path"
    - "Scaffolded main.go after init contains clear TODO guidance for server wiring, not a useless print+exit stub"
    - "forge dev --help clearly explains it is a code regeneration watcher, not an app runner"
  artifacts:
    - path: "internal/scaffold/scaffold.go"
      provides: "InferModule returns just the project name"
      contains: "return projectName"
    - path: "internal/scaffold/templates/main.go.tmpl"
      provides: "Useful main.go scaffold with server setup guidance"
      contains: "TODO"
    - path: "internal/cli/dev.go"
      provides: "Clear help text for forge dev"
      contains: "regenerat"
  key_links:
    - from: "internal/cli/init.go"
      to: "internal/scaffold/scaffold.go"
      via: "scaffold.InferModule(projectName)"
      pattern: "InferModule"
---

<objective>
Fix three forge init/generate UX issues: hardcoded GitHub module path, useless main.go stub, and confusing forge dev help text.

Purpose: New users running `forge init` get a working, understandable starting point instead of a broken module path, an empty main.go, and confusing CLI help.
Output: Updated scaffold logic, main.go template, and dev command help text.
</objective>

<execution_context>
@/Users/nathananderson-tennant/.claude/get-shit-done/workflows/execute-plan.md
@/Users/nathananderson-tennant/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@internal/scaffold/scaffold.go
@internal/scaffold/templates/main.go.tmpl
@internal/scaffold/templates/go.mod.tmpl
@internal/scaffold/templates/forge.toml.tmpl
@internal/cli/init.go
@internal/cli/dev.go
@internal/watcher/dev.go
@internal/generator/templates/api_register_all.go.tmpl
</context>

<tasks>

<task type="auto">
  <name>Task 1: Fix module path inference and main.go scaffold template</name>
  <files>
    internal/scaffold/scaffold.go
    internal/scaffold/templates/main.go.tmpl
  </files>
  <action>
**Fix InferModule in scaffold.go:**

Replace the `InferModule` function to simply return the project name as-is. Remove the git config detection logic entirely. The generated project belongs to the USER, not the forge developer. The user's git username has nothing to do with their desired Go module path. A bare project name (e.g., `myapp`) is a valid Go module path and is the correct default for a new project that doesn't have a GitHub repo yet.

The function should become:

```go
func InferModule(projectName string) string {
	return projectName
}
```

**Fix main.go.tmpl:**

Replace the current useless stub (print + os.Exit) with a main.go that provides clear scaffolding direction. The generated code puts API routes in `gen/api/` and HTML routes in `gen/html/`, wired via `RegisterAllRoutes` and `RegisterAllHTMLRoutes`. Since the user hasn't run `forge generate` yet at init time, main.go can't import gen/ packages. Instead, provide a well-commented skeleton that shows:

1. A `main()` that calls `run()` returning error (standard Go pattern)
2. Inside `run()`, comments explaining the next steps after `forge generate`
3. A commented-out example of what the server wiring will look like (chi router + huma API)
4. A working `fmt.Println` that tells the user to run `forge generate` first

Keep it practical (under 40 lines). Use the `{{.Module}}` template variable in the commented import paths so they match the actual module. Do NOT import packages that don't exist yet -- the code must compile as-is after `forge init`.

The template should look approximately like:

```
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Println("{{.Name}} — powered by Forge")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Define resources in resources/*/schema.go")
	fmt.Println("  2. Run 'forge generate' to generate code")
	fmt.Println("  3. Wire up your server here in main.go")
	fmt.Println()
	fmt.Println("After generating, your API routes will be in gen/api/")
	fmt.Println("and HTML routes in gen/html/. See the README for details.")
	return nil
}
```

This is immediately useful: it compiles, it tells the user what to do next, and the `run() error` pattern is the correct foundation they'll build on.
  </action>
  <verify>
    Run `go build ./internal/scaffold/...` to confirm compilation.
    Run `go test ./internal/scaffold/...` if tests exist.
    Visually inspect that InferModule just returns projectName.
    Visually inspect that main.go.tmpl compiles standalone and provides clear next-step guidance.
  </verify>
  <done>
    InferModule("myapp") returns "myapp" (not "github.com/anyone/myapp").
    main.go.tmpl produces a compilable Go file with run() error pattern and clear next-steps guidance.
  </done>
</task>

<task type="auto">
  <name>Task 2: Fix forge dev help text to clarify it is a code regeneration watcher</name>
  <files>
    internal/cli/dev.go
  </files>
  <action>
Update the `newDevCmd()` function in dev.go to have clearer Short and Long descriptions.

Current Short: "Start development server with file watching"
Problem: "development server" implies it runs your app (like `npm run dev`). It does NOT run the app -- it only watches files and re-runs `forge generate` on changes.

New Short: "Watch files and regenerate code on changes"

Current Long: "Starts a development server that watches for file changes..."
Problem: Same "development server" confusion, plus vague about what actually happens.

New Long should clearly explain:
- It watches `resources/` and `internal/` for file changes
- When changes are detected, it automatically re-runs code generation (equivalent to `forge generate`)
- It does NOT run your application -- use `go run .` for that
- Mention the watched file types: .go, .templ, .sql, .css
- Keep the "Ctrl+C to stop" note

Something like:

```
Watches for file changes in resources/ and internal/ directories and
automatically re-runs code generation when files change. This is equivalent
to running 'forge generate' each time you save a file.

This command does NOT run your application. To run your app, use 'go run .'
in a separate terminal.

Watched file types: .go, .templ, .sql, .css

Ctrl+C to stop.
```
  </action>
  <verify>
    Run `go build ./internal/cli/...` to confirm compilation.
    Run `go run . dev --help` (from project root with the built binary, or `go run ./cmd/forge dev --help`) to confirm the new help text renders correctly.
  </verify>
  <done>
    `forge dev --help` clearly states it is a file watcher that regenerates code, explicitly notes it does NOT run the application, and tells users to use `go run .` separately.
  </done>
</task>

</tasks>

<verification>
1. `go build ./...` passes (all modified packages compile)
2. `go test ./...` passes (no regressions)
3. InferModule returns bare project name without github.com prefix
4. main.go.tmpl produces valid, compilable Go with clear next-step instructions
5. `forge dev --help` output is unambiguous about what the command does
</verification>

<success_criteria>
- forge init myapp creates go.mod with `module myapp`
- forge init myapp creates a main.go with run() pattern and next-steps guidance, not a print+exit stub
- forge dev --help clearly describes a file watcher for code regeneration, not an app server
- All existing tests pass
</success_criteria>

<output>
After completion, create `.planning/quick/1-fix-forge-init-ux-module-path-main-go-sc/1-SUMMARY.md`
</output>
