package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/alternayte/forge/internal/config"
	"github.com/alternayte/forge/internal/parser"
	"github.com/spf13/cobra"

	"path/filepath"
)

func newOpenapiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openapi",
		Short: "OpenAPI spec management",
	}
	return cmd
}

func newOpenapiExportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [file]",
		Short: "Export the OpenAPI spec to a file",
		Long:  "Export the OpenAPI 3.1 spec generated from your schemas. Supports JSON and YAML formats.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runOpenapiExport,
	}
	cmd.Flags().String("format", "json", "Output format: json or yaml")
	return cmd
}

func runOpenapiExport(cmd *cobra.Command, args []string) error {
	format, _ := cmd.Flags().GetString("format")
	if format != "json" && format != "yaml" {
		return fmt.Errorf("unsupported format %q: must be 'json' or 'yaml'", format)
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
	}

	// Load config
	cfg, err := config.Load(filepath.Join(projectRoot, "forge.toml"))
	if err != nil {
		return fmt.Errorf("failed to load forge.toml: %w", err)
	}

	// Parse schemas
	resourcesDir := filepath.Join(projectRoot, "resources")
	result, err := parser.ParseDir(resourcesDir)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %w", err)
	}

	if len(result.Resources) == 0 {
		fmt.Println()
		fmt.Println("  No schema definitions found in resources/. Create a schema file first.")
		fmt.Println()
		return nil
	}

	// Build spec bytes from IR
	specBytes, err := buildSpecFromIR(result.Resources, cfg.Project.Module, format)
	if err != nil {
		return fmt.Errorf("failed to build OpenAPI spec: %w", err)
	}

	// Determine output filename
	var outputFile string
	if len(args) > 0 {
		outputFile = args[0]
	} else {
		outputFile = fmt.Sprintf("openapi.%s", format)
	}

	if err := os.WriteFile(outputFile, specBytes, 0644); err != nil {
		return fmt.Errorf("failed to write spec to %s: %w", outputFile, err)
	}

	fmt.Printf("\n  OpenAPI spec exported to %s\n\n", outputFile)
	return nil
}

// buildSpecFromIR generates an OpenAPI 3.1 spec from parsed resource IR.
// It builds the huma.OpenAPI struct directly from IR data, populating paths
// with the same 5 CRUD operations per resource that forge generates.
// Every operation has an explicit operationId, tags, and summary for SDK readiness.
func buildSpecFromIR(resources []parser.ResourceIR, projectModule string, format string) ([]byte, error) {
	// Build project name from module path (last segment)
	projectName := "Forge API"
	if projectModule != "" {
		parts := splitModulePath(projectModule)
		if len(parts) > 0 {
			projectName = capitalizeFirst(parts[len(parts)-1]) + " API"
		}
	}

	// Construct the OpenAPI document
	spec := &huma.OpenAPI{
		OpenAPI: "3.1.0",
		Info: &huma.Info{
			Title:   projectName,
			Version: "1.0.0",
		},
		Paths: map[string]*huma.PathItem{},
	}

	// Collect unique tag names for top-level tags section
	tagNames := make(map[string]bool)

	// Add operations for each resource
	for _, resource := range resources {
		name := resource.Name
		pluralName := routePlural(name)
		kebabPlural := routeKebab(pluralName)
		tag := name

		tagNames[tag] = true

		collectionPath := fmt.Sprintf("/api/v1/%s", kebabPlural)
		itemPath := fmt.Sprintf("/api/v1/%s/{id}", kebabPlural)

		// Ensure path items exist
		if spec.Paths[collectionPath] == nil {
			spec.Paths[collectionPath] = &huma.PathItem{}
		}
		if spec.Paths[itemPath] == nil {
			spec.Paths[itemPath] = &huma.PathItem{}
		}

		// Build operations and add via AddOperation (handles Paths map population)
		ops := resourceOperations(name, pluralName, kebabPlural, tag)
		for _, op := range ops {
			spec.AddOperation(op)
		}
	}

	// Add top-level tags (one per resource)
	for tagName := range tagNames {
		spec.Tags = append(spec.Tags, &huma.Tag{
			Name:        tagName,
			Description: fmt.Sprintf("%s resource endpoints", tagName),
		})
	}

	// Marshal to requested format
	switch format {
	case "yaml":
		return spec.YAML()
	default:
		return json.MarshalIndent(spec, "", "  ")
	}
}

// resourceOperations returns the 5 CRUD huma.Operation objects for a resource.
// Every operation has OperationID, Tags, Summary, Method, and Path set.
func resourceOperations(name, pluralName, kebabPlural, tag string) []*huma.Operation {
	collectionPath := fmt.Sprintf("/api/v1/%s", kebabPlural)
	itemPath := fmt.Sprintf("/api/v1/%s/{id}", kebabPlural)

	idParam := &huma.Param{
		Name:     "id",
		In:       "path",
		Required: true,
		Schema: &huma.Schema{
			Type:   "string",
			Format: "uuid",
		},
	}

	return []*huma.Operation{
		{
			OperationID: fmt.Sprintf("list%s", pluralName),
			Method:      "GET",
			Path:        collectionPath,
			Tags:        []string{tag},
			Summary:     fmt.Sprintf("List %s", pluralName),
			Parameters: []*huma.Param{
				{Name: "cursor", In: "query", Schema: &huma.Schema{Type: "string"}},
				{Name: "limit", In: "query", Schema: &huma.Schema{Type: "integer", Default: json.RawMessage("50")}},
				{Name: "sort", In: "query", Schema: &huma.Schema{Type: "string"}},
				{Name: "order", In: "query", Schema: &huma.Schema{Type: "string", Enum: []any{"asc", "desc"}}},
			},
			Responses: map[string]*huma.Response{
				"200": {
					Description: fmt.Sprintf("List of %s", pluralName),
					Content: map[string]*huma.MediaType{
						"application/json": {
							Schema: &huma.Schema{
								Type: "object",
								Properties: map[string]*huma.Schema{
									"data": {
										Type: "array",
										Items: &huma.Schema{
											Type: "object",
										},
									},
									"pagination": {
										Type: "object",
										Properties: map[string]*huma.Schema{
											"cursor":     {Type: "string"},
											"has_more":   {Type: "boolean"},
											"limit":      {Type: "integer"},
											"total_count": {Type: "integer"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			OperationID: fmt.Sprintf("get%s", name),
			Method:      "GET",
			Path:        itemPath,
			Tags:        []string{tag},
			Summary:     fmt.Sprintf("Get a %s by ID", name),
			Parameters:  []*huma.Param{idParam},
			Responses: map[string]*huma.Response{
				"200": {
					Description: fmt.Sprintf("The requested %s", name),
					Content: map[string]*huma.MediaType{
						"application/json": {
							Schema: &huma.Schema{
								Type: "object",
								Properties: map[string]*huma.Schema{
									"data": {Type: "object"},
								},
							},
						},
					},
				},
				"404": {Description: fmt.Sprintf("%s not found", name)},
			},
		},
		{
			OperationID: fmt.Sprintf("create%s", name),
			Method:      "POST",
			Path:        collectionPath,
			Tags:        []string{tag},
			Summary:     fmt.Sprintf("Create a new %s", name),
			RequestBody: &huma.RequestBody{
				Required: true,
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: &huma.Schema{Type: "object"},
					},
				},
			},
			Responses: map[string]*huma.Response{
				"201": {
					Description: fmt.Sprintf("%s created successfully", name),
					Content: map[string]*huma.MediaType{
						"application/json": {
							Schema: &huma.Schema{
								Type: "object",
								Properties: map[string]*huma.Schema{
									"data": {Type: "object"},
								},
							},
						},
					},
				},
				"422": {Description: "Validation error"},
			},
		},
		{
			OperationID: fmt.Sprintf("update%s", name),
			Method:      "PUT",
			Path:        itemPath,
			Tags:        []string{tag},
			Summary:     fmt.Sprintf("Update an existing %s", name),
			Parameters:  []*huma.Param{idParam},
			RequestBody: &huma.RequestBody{
				Required: true,
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: &huma.Schema{Type: "object"},
					},
				},
			},
			Responses: map[string]*huma.Response{
				"200": {
					Description: fmt.Sprintf("%s updated successfully", name),
					Content: map[string]*huma.MediaType{
						"application/json": {
							Schema: &huma.Schema{
								Type: "object",
								Properties: map[string]*huma.Schema{
									"data": {Type: "object"},
								},
							},
						},
					},
				},
				"404": {Description: fmt.Sprintf("%s not found", name)},
				"422": {Description: "Validation error"},
			},
		},
		{
			OperationID: fmt.Sprintf("delete%s", name),
			Method:      "DELETE",
			Path:        itemPath,
			Tags:        []string{tag},
			Summary:     fmt.Sprintf("Delete a %s", name),
			Parameters:  []*huma.Param{idParam},
			Responses: map[string]*huma.Response{
				"204": {Description: fmt.Sprintf("%s deleted successfully", name)},
				"404": {Description: fmt.Sprintf("%s not found", name)},
			},
		},
	}
}

// splitModulePath splits a Go module path on '/' separators.
func splitModulePath(module string) []string {
	var parts []string
	current := ""
	for _, c := range module {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
			}
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// capitalizeFirst returns s with the first character uppercased.
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	if runes[0] >= 'a' && runes[0] <= 'z' {
		runes[0] = runes[0] - 'a' + 'A'
	}
	return string(runes)
}
