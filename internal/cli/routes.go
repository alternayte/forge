package cli

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/forge-framework/forge/internal/parser"
	"github.com/spf13/cobra"
)

// Route represents a single API route entry.
type Route struct {
	Method      string
	Path        string
	OperationID string
}

func newRoutesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routes",
		Short: "List all registered routes",
		Long:  "Display all HTML and API routes grouped by resource",
		RunE:  runRoutes,
	}
	return cmd
}

func runRoutes(cmd *cobra.Command, args []string) error {
	// Find project root by looking for forge.toml
	projectRoot, err := findProjectRoot()
	if err != nil {
		return fmt.Errorf("not a forge project (forge.toml not found). Run 'forge init' first")
	}

	// Parse schemas from resources/ directory
	resourcesDir := fmt.Sprintf("%s/resources", projectRoot)
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

	// Collect all routes grouped by resource (both API and HTML)
	type resourceRouteGroup struct {
		Name      string
		APIRoutes []Route
		HTMLRoutes []Route
	}

	var groups []resourceRouteGroup
	totalRoutes := 0

	for _, resource := range result.Resources {
		api := apiRoutes(resource)
		html := htmlRoutes(resource)
		groups = append(groups, resourceRouteGroup{
			Name:       resource.Name,
			APIRoutes:  api,
			HTMLRoutes: html,
		})
		totalRoutes += len(api) + len(html)
	}

	// Compute max path length across all routes for column alignment
	maxPathLen := 0
	for _, g := range groups {
		for _, r := range g.APIRoutes {
			if len(r.Path) > maxPathLen {
				maxPathLen = len(r.Path)
			}
		}
		for _, r := range g.HTMLRoutes {
			if len(r.Path) > maxPathLen {
				maxPathLen = len(r.Path)
			}
		}
	}

	// Define lipgloss styles
	headerStyle := lipgloss.NewStyle().Bold(true)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7")) // dim white
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	getStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))   // green
	postStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))  // blue
	putStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))   // yellow
	deleteStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red

	// printRoute prints a single route row with color-coded method and aligned path.
	printRoute := func(route Route) {
		methodStr := fmt.Sprintf("%-6s", route.Method)
		var coloredMethod string
		switch route.Method {
		case "GET":
			coloredMethod = getStyle.Render(methodStr)
		case "POST":
			coloredMethod = postStyle.Render(methodStr)
		case "PUT":
			coloredMethod = putStyle.Render(methodStr)
		case "DELETE":
			coloredMethod = deleteStyle.Render(methodStr)
		default:
			coloredMethod = methodStr
		}

		paddedPath := fmt.Sprintf("%-*s", maxPathLen+2, route.Path)
		fmt.Printf("      %s %s%s\n",
			coloredMethod,
			paddedPath,
			dimStyle.Render(route.OperationID),
		)
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("Routes:"))
	fmt.Println()

	for _, g := range groups {
		totalResourceRoutes := len(g.APIRoutes) + len(g.HTMLRoutes)
		suffix := "routes"
		if totalResourceRoutes == 1 {
			suffix = "route"
		}
		fmt.Printf("  %s\n", headerStyle.Render(fmt.Sprintf("%s (%d %s)", g.Name, totalResourceRoutes, suffix)))
		fmt.Println()

		// API Routes section
		fmt.Printf("    %s\n", sectionStyle.Render("API"))
		for _, route := range g.APIRoutes {
			printRoute(route)
		}
		fmt.Println()

		// HTML Routes section
		fmt.Printf("    %s\n", sectionStyle.Render("HTML"))
		for _, route := range g.HTMLRoutes {
			printRoute(route)
		}
		fmt.Println()
	}

	// Total summary
	resourceWord := "resources"
	if len(groups) == 1 {
		resourceWord = "resource"
	}
	fmt.Printf("  Total: %d %s (%d %s)\n",
		totalRoutes,
		"routes",
		len(groups),
		resourceWord,
	)
	fmt.Println()

	return nil
}

// apiRoutes returns the 5 standard CRUD API routes for a resource.
// Structured as a separate function so HTML routes can be added alongside in Phase 6.
func apiRoutes(resource parser.ResourceIR) []Route {
	name := resource.Name
	pluralName := routePlural(name)
	kebabPlural := routeKebab(pluralName)

	return []Route{
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/api/v1/%s", kebabPlural),
			OperationID: fmt.Sprintf("list%s", pluralName),
		},
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/api/v1/%s/{id}", kebabPlural),
			OperationID: fmt.Sprintf("get%s", name),
		},
		{
			Method:      "POST",
			Path:        fmt.Sprintf("/api/v1/%s", kebabPlural),
			OperationID: fmt.Sprintf("create%s", name),
		},
		{
			Method:      "PUT",
			Path:        fmt.Sprintf("/api/v1/%s/{id}", kebabPlural),
			OperationID: fmt.Sprintf("update%s", name),
		},
		{
			Method:      "DELETE",
			Path:        fmt.Sprintf("/api/v1/%s/{id}", kebabPlural),
			OperationID: fmt.Sprintf("delete%s", name),
		},
	}
}

// htmlRoutes returns the 7 HTML CRUD routes for a resource.
// HTML routes use root path patterns (no /api/v1/ prefix) and are served via
// Datastar SSE for create/update/delete mutations.
func htmlRoutes(resource parser.ResourceIR) []Route {
	name := resource.Name
	pluralName := routePlural(name)
	kebabPlural := routeKebab(pluralName)

	return []Route{
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/%s", kebabPlural),
			OperationID: fmt.Sprintf("html.list%s", pluralName),
		},
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/%s/new", kebabPlural),
			OperationID: fmt.Sprintf("html.new%s", name),
		},
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/%s/{id}", kebabPlural),
			OperationID: fmt.Sprintf("html.get%s", name),
		},
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/%s/{id}/edit", kebabPlural),
			OperationID: fmt.Sprintf("html.edit%s", name),
		},
		{
			Method:      "POST",
			Path:        fmt.Sprintf("/%s", kebabPlural),
			OperationID: fmt.Sprintf("html.create%s", name),
		},
		{
			Method:      "PUT",
			Path:        fmt.Sprintf("/%s/{id}", kebabPlural),
			OperationID: fmt.Sprintf("html.update%s", name),
		},
		{
			Method:      "DELETE",
			Path:        fmt.Sprintf("/%s/{id}", kebabPlural),
			OperationID: fmt.Sprintf("html.delete%s", name),
		},
	}
}

// routePlural naively pluralizes a resource name.
func routePlural(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

// routeKebab converts PascalCase to kebab-case (e.g., "BlogPost" -> "blog-post").
func routeKebab(s string) string {
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

// routeLowerCamel converts PascalCase to camelCase.
func routeLowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	return strings.ToLower(string(runes[0:1])) + string(runes[1:])
}
