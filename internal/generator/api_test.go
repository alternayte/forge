package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

func TestGenerateAPI(t *testing.T) {
	// Create temp directory for test output
	tempDir := t.TempDir()

	// Define test resource: Product with Name/Required/MaxLen, Price/Decimal/Required, Status/Enum/Filterable
	resource := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{
				Name: "Name",
				Type: "String",
				Modifiers: []parser.ModifierIR{
					{Type: "Required"},
					{Type: "MaxLen", Value: 200},
				},
			},
			{
				Name: "Price",
				Type: "Decimal",
				Modifiers: []parser.ModifierIR{
					{Type: "Required"},
				},
			},
			{
				Name:       "Status",
				Type:       "Enum",
				EnumValues: []string{"active", "inactive", "draft"},
				Modifiers: []parser.ModifierIR{
					{Type: "Filterable"},
				},
			},
		},
	}

	// Call GenerateAPI
	err := GenerateAPI([]parser.ResourceIR{resource}, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateAPI failed: %v", err)
	}

	// --- Verify types.go ---
	typesPath := filepath.Join(tempDir, "api", "types.go")
	typesContent, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("Failed to read generated types.go: %v", err)
	}
	typesStr := string(typesContent)

	if !strings.Contains(typesStr, "PaginationMeta") {
		t.Error("Generated types.go missing PaginationMeta struct")
	}
	if !strings.Contains(typesStr, "toHumaError") {
		t.Error("Generated types.go missing toHumaError function")
	}
	if !strings.Contains(typesStr, `json:"page"`) {
		t.Error("Generated types.go PaginationMeta missing page field")
	}
	if !strings.Contains(typesStr, `json:"has_more"`) {
		t.Error("Generated types.go PaginationMeta missing has_more field")
	}

	// --- Verify inputs.go ---
	inputsPath := filepath.Join(tempDir, "api", "product_inputs.go")
	inputsContent, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatalf("Failed to read generated product_inputs.go: %v", err)
	}
	inputsStr := string(inputsContent)

	// ListProductInput with Cursor, Limit, Status filter param
	if !strings.Contains(inputsStr, "ListProductInput") {
		t.Error("Generated product_inputs.go missing ListProductInput")
	}
	if !strings.Contains(inputsStr, `query:"cursor"`) {
		t.Error("Generated product_inputs.go ListProductInput missing cursor query param")
	}
	if !strings.Contains(inputsStr, `query:"limit"`) {
		t.Error("Generated product_inputs.go ListProductInput missing limit query param")
	}
	if !strings.Contains(inputsStr, `query:"status"`) {
		t.Error("Generated product_inputs.go ListProductInput missing status filter param")
	}

	// CreateProductInput with Body.Name (minLength, maxLength tags)
	if !strings.Contains(inputsStr, "CreateProductInput") {
		t.Error("Generated product_inputs.go missing CreateProductInput")
	}
	if !strings.Contains(inputsStr, `maxLength:"200"`) {
		t.Error("Generated product_inputs.go CreateProductInput missing maxLength tag from MaxLen modifier")
	}
	if !strings.Contains(inputsStr, "Name") {
		t.Error("Generated product_inputs.go CreateProductInput missing Name field")
	}
	if !strings.Contains(inputsStr, "Price") {
		t.Error("Generated product_inputs.go CreateProductInput missing Price field")
	}

	// UpdateProductInput
	if !strings.Contains(inputsStr, "UpdateProductInput") {
		t.Error("Generated product_inputs.go missing UpdateProductInput")
	}
	if !strings.Contains(inputsStr, `path:"id"`) {
		t.Error("Generated product_inputs.go missing path:\"id\" tag")
	}

	// GetProductInput, DeleteProductInput
	if !strings.Contains(inputsStr, "GetProductInput") {
		t.Error("Generated product_inputs.go missing GetProductInput")
	}
	if !strings.Contains(inputsStr, "DeleteProductInput") {
		t.Error("Generated product_inputs.go missing DeleteProductInput")
	}

	// --- Verify outputs.go ---
	outputsPath := filepath.Join(tempDir, "api", "product_outputs.go")
	outputsContent, err := os.ReadFile(outputsPath)
	if err != nil {
		t.Fatalf("Failed to read generated product_outputs.go: %v", err)
	}
	outputsStr := string(outputsContent)

	// ListProductOutput with Data []Product and PaginationMeta
	if !strings.Contains(outputsStr, "ListProductOutput") {
		t.Error("Generated product_outputs.go missing ListProductOutput")
	}
	if !strings.Contains(outputsStr, "[]models.Product") {
		t.Error("Generated product_outputs.go ListProductOutput missing Data []Product field")
	}
	if !strings.Contains(outputsStr, "PaginationMeta") {
		t.Error("Generated product_outputs.go ListProductOutput missing PaginationMeta field")
	}
	if !strings.Contains(outputsStr, `json:"data"`) {
		t.Error("Generated product_outputs.go missing json:\"data\" tag")
	}
	if !strings.Contains(outputsStr, `json:"pagination"`) {
		t.Error("Generated product_outputs.go missing json:\"pagination\" tag")
	}

	// GetProductOutput with Data Product wrapped in json:"data"
	if !strings.Contains(outputsStr, "GetProductOutput") {
		t.Error("Generated product_outputs.go missing GetProductOutput")
	}
	if !strings.Contains(outputsStr, "models.Product") {
		t.Error("Generated product_outputs.go missing models.Product reference")
	}

	// CreateProductOutput, UpdateProductOutput, DeleteProductOutput
	if !strings.Contains(outputsStr, "CreateProductOutput") {
		t.Error("Generated product_outputs.go missing CreateProductOutput")
	}
	if !strings.Contains(outputsStr, "UpdateProductOutput") {
		t.Error("Generated product_outputs.go missing UpdateProductOutput")
	}
	if !strings.Contains(outputsStr, "DeleteProductOutput") {
		t.Error("Generated product_outputs.go missing DeleteProductOutput")
	}

	// Link header field on list output
	if !strings.Contains(outputsStr, `header:"Link"`) {
		t.Error("Generated product_outputs.go ListProductOutput missing Link header field")
	}

	// --- Verify routes.go ---
	routesPath := filepath.Join(tempDir, "api", "product_routes.go")
	routesContent, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("Failed to read generated product_routes.go: %v", err)
	}
	routesStr := string(routesContent)

	// RegisterProductRoutes function
	if !strings.Contains(routesStr, "RegisterProductRoutes") {
		t.Error("Generated product_routes.go missing RegisterProductRoutes function")
	}

	// huma.Register calls with correct OperationIDs
	if !strings.Contains(routesStr, `"listProducts"`) {
		t.Error("Generated product_routes.go missing listProducts operationId")
	}
	if !strings.Contains(routesStr, `"getProduct"`) {
		t.Error("Generated product_routes.go missing getProduct operationId")
	}
	if !strings.Contains(routesStr, `"createProduct"`) {
		t.Error("Generated product_routes.go missing createProduct operationId")
	}
	if !strings.Contains(routesStr, `"updateProduct"`) {
		t.Error("Generated product_routes.go missing updateProduct operationId")
	}
	if !strings.Contains(routesStr, `"deleteProduct"`) {
		t.Error("Generated product_routes.go missing deleteProduct operationId")
	}

	// Correct /api/v1/products path
	if !strings.Contains(routesStr, "/api/v1/products") {
		t.Error("Generated product_routes.go missing /api/v1/products path")
	}

	// Calls to action layer
	if !strings.Contains(routesStr, "act.List") {
		t.Error("Generated product_routes.go missing act.List call")
	}
	if !strings.Contains(routesStr, "act.Get") {
		t.Error("Generated product_routes.go missing act.Get call")
	}
	if !strings.Contains(routesStr, "act.Create") {
		t.Error("Generated product_routes.go missing act.Create call")
	}
	if !strings.Contains(routesStr, "act.Update") {
		t.Error("Generated product_routes.go missing act.Update call")
	}
	if !strings.Contains(routesStr, "act.Delete") {
		t.Error("Generated product_routes.go missing act.Delete call")
	}

	// huma.Register calls present
	if !strings.Contains(routesStr, "huma.Register") {
		t.Error("Generated product_routes.go missing huma.Register calls")
	}

	// Link header setting logic in List handler (RFC 8288 rel="next")
	if !strings.Contains(routesStr, `rel="next"`) {
		t.Error("Generated product_routes.go List handler missing RFC 8288 rel=\"next\" Link header")
	}
	if !strings.Contains(routesStr, "out.Link") {
		t.Error("Generated product_routes.go List handler missing out.Link assignment")
	}
	if !strings.Contains(routesStr, "buildAPILinkHeader") {
		t.Error("Generated product_routes.go List handler missing buildAPILinkHeader call")
	}
}

func TestGenerateAPI_MultipleResources(t *testing.T) {
	// Create temp directory for test output
	tempDir := t.TempDir()

	// Define two test resources
	resources := []parser.ResourceIR{
		{
			Name: "Product",
			Fields: []parser.FieldIR{
				{Name: "ID", Type: "UUID"},
				{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			},
		},
		{
			Name: "Category",
			Fields: []parser.FieldIR{
				{Name: "ID", Type: "UUID"},
				{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			},
		},
	}

	// Call GenerateAPI
	err := GenerateAPI(resources, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateAPI failed: %v", err)
	}

	// Verify types.go exists once
	typesPath := filepath.Join(tempDir, "api", "types.go")
	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Error("types.go was not generated")
	}

	// Verify product files exist
	for _, suffix := range []string{"_inputs.go", "_outputs.go", "_routes.go"} {
		p := filepath.Join(tempDir, "api", "product"+suffix)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("product%s was not generated", suffix)
		}
	}

	// Verify category files exist
	for _, suffix := range []string{"_inputs.go", "_outputs.go", "_routes.go"} {
		p := filepath.Join(tempDir, "api", "category"+suffix)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("category%s was not generated", suffix)
		}
	}

	// Verify category routes reference CategoryActions
	categoryRoutesPath := filepath.Join(tempDir, "api", "category_routes.go")
	categoryRoutesContent, err := os.ReadFile(categoryRoutesPath)
	if err != nil {
		t.Fatalf("Failed to read generated category_routes.go: %v", err)
	}
	categoryRoutesStr := string(categoryRoutesContent)

	if !strings.Contains(categoryRoutesStr, "RegisterCategoryRoutes") {
		t.Error("Generated category_routes.go missing RegisterCategoryRoutes function")
	}
	if !strings.Contains(categoryRoutesStr, "CategoryActions") {
		t.Error("Generated category_routes.go missing CategoryActions reference")
	}
	if !strings.Contains(categoryRoutesStr, "/api/v1/categories") {
		t.Error("Generated category_routes.go missing /api/v1/categories path")
	}
}

func TestGenerateAPI_OperationIDs(t *testing.T) {
	// Create temp directory for test output
	tempDir := t.TempDir()

	// BlogPost resource (multi-word to test camelCase conversion)
	resource := parser.ResourceIR{
		Name: "BlogPost",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Title", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
		},
	}

	err := GenerateAPI([]parser.ResourceIR{resource}, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateAPI failed: %v", err)
	}

	routesPath := filepath.Join(tempDir, "api", "blog_post_routes.go")
	routesContent, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("Failed to read generated blog_post_routes.go: %v", err)
	}
	routesStr := string(routesContent)

	// Verify camelCase verb+noun operationIds
	if !strings.Contains(routesStr, `"listBlogPosts"`) {
		t.Errorf("Generated routes.go missing listBlogPosts operationId, got:\n%s", routesStr)
	}
	if !strings.Contains(routesStr, `"getBlogPost"`) {
		t.Errorf("Generated routes.go missing getBlogPost operationId")
	}
	if !strings.Contains(routesStr, `"createBlogPost"`) {
		t.Errorf("Generated routes.go missing createBlogPost operationId")
	}
	if !strings.Contains(routesStr, `"updateBlogPost"`) {
		t.Errorf("Generated routes.go missing updateBlogPost operationId")
	}
	if !strings.Contains(routesStr, `"deleteBlogPost"`) {
		t.Errorf("Generated routes.go missing deleteBlogPost operationId")
	}

	// Verify kebab-case URL paths
	if !strings.Contains(routesStr, "/api/v1/blog-posts") {
		t.Errorf("Generated routes.go missing /api/v1/blog-posts path (kebab-case)")
	}
}

func TestGenerateAPI_LinkHeader(t *testing.T) {
	// Create temp directory for test output
	tempDir := t.TempDir()

	resource := parser.ResourceIR{
		Name: "Article",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Title", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
		},
	}

	err := GenerateAPI([]parser.ResourceIR{resource}, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateAPI failed: %v", err)
	}

	routesPath := filepath.Join(tempDir, "api", "article_routes.go")
	routesContent, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("Failed to read generated article_routes.go: %v", err)
	}
	routesStr := string(routesContent)

	// Verify RFC 8288 Link header generation in List handler
	if !strings.Contains(routesStr, `rel="next"`) {
		t.Error("Generated List handler missing RFC 8288 rel=\"next\" in Link header")
	}
	if !strings.Contains(routesStr, "hasMore") {
		t.Error("Generated List handler missing hasMore condition for Link header")
	}
	if !strings.Contains(routesStr, "out.Link") {
		t.Error("Generated List handler missing out.Link assignment (header field)")
	}
	if !strings.Contains(routesStr, "buildAPILinkHeader") {
		t.Error("Generated List handler missing buildAPILinkHeader call")
	}
}
