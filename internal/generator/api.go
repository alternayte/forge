package generator

import (
	"path/filepath"

	"github.com/forge-framework/forge/internal/parser"
)

// GenerateAPI generates Huma API Input/Output structs and route registration functions for all resources.
func GenerateAPI(resources []parser.ResourceIR, outputDir string, projectModule string) error {
	// Create api directory
	apiDir := filepath.Join(outputDir, "api")
	if err := ensureDir(apiDir); err != nil {
		return err
	}

	// Generate shared types.go file first (PaginationMeta, toHumaError, buildAPILinkHeader)
	typesData := struct {
		ProjectModule string
	}{
		ProjectModule: projectModule,
	}

	typesRaw, err := renderTemplate("templates/api_types.go.tmpl", typesData)
	if err != nil {
		return err
	}

	typesPath := filepath.Join(apiDir, "types.go")
	if err := writeGoFile(typesPath, typesRaw); err != nil {
		return err
	}

	// Generate per-resource files
	for _, resource := range resources {
		// Prepare template data with ProjectModule
		data := struct {
			parser.ResourceIR
			ProjectModule string
		}{
			ResourceIR:    resource,
			ProjectModule: projectModule,
		}

		// Render api_inputs.go.tmpl -> gen/api/{snake}_inputs.go
		inputsRaw, err := renderTemplate("templates/api_inputs.go.tmpl", data)
		if err != nil {
			return err
		}
		inputsPath := filepath.Join(apiDir, snake(resource.Name)+"_inputs.go")
		if err := writeGoFile(inputsPath, inputsRaw); err != nil {
			return err
		}

		// Render api_outputs.go.tmpl -> gen/api/{snake}_outputs.go
		outputsRaw, err := renderTemplate("templates/api_outputs.go.tmpl", data)
		if err != nil {
			return err
		}
		outputsPath := filepath.Join(apiDir, snake(resource.Name)+"_outputs.go")
		if err := writeGoFile(outputsPath, outputsRaw); err != nil {
			return err
		}

		// Render api_register.go.tmpl -> gen/api/{snake}_routes.go
		routesRaw, err := renderTemplate("templates/api_register.go.tmpl", data)
		if err != nil {
			return err
		}
		routesPath := filepath.Join(apiDir, snake(resource.Name)+"_routes.go")
		if err := writeGoFile(routesPath, routesRaw); err != nil {
			return err
		}
	}

	// Generate shared register_all.go (RegisterAllRoutes dispatcher over all resources)
	registerAllData := struct {
		ProjectModule string
		Resources     []parser.ResourceIR
	}{
		ProjectModule: projectModule,
		Resources:     resources,
	}

	registerAllRaw, err := renderTemplate("templates/api_register_all.go.tmpl", registerAllData)
	if err != nil {
		return err
	}

	registerAllPath := filepath.Join(apiDir, "register_all.go")
	if err := writeGoFile(registerAllPath, registerAllRaw); err != nil {
		return err
	}

	return nil
}
