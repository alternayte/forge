package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alternayte/forge/internal/parser"
)

func TestGenerateActions(t *testing.T) {
	// Create temp directory for test output
	tempDir := t.TempDir()

	// Define test resource matching pattern from other tests
	resource := parser.ResourceIR{
		Name: "Product",
		Fields: []parser.FieldIR{
			{Name: "ID", Type: "UUID"},
			{Name: "Name", Type: "String", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			{Name: "Price", Type: "Decimal", Modifiers: []parser.ModifierIR{{Type: "Required"}}},
			{Name: "Description", Type: "Text"},
		},
	}

	// Call GenerateActions
	err := GenerateActions([]parser.ResourceIR{resource}, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateActions failed: %v", err)
	}

	// Verify types.go was generated
	typesPath := filepath.Join(tempDir, "actions", "types.go")
	typesContent, err := os.ReadFile(typesPath)
	if err != nil {
		t.Fatalf("Failed to read generated types.go: %v", err)
	}

	typesStr := string(typesContent)

	// Check for DB interface
	if !strings.Contains(typesStr, "type DB interface") {
		t.Error("Generated types.go missing 'type DB interface'")
	}
	if !strings.Contains(typesStr, "QueryRow(ctx context.Context, sql string, args ...any) pgx.Row") {
		t.Error("Generated types.go missing QueryRow method in DB interface")
	}
	if !strings.Contains(typesStr, "Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)") {
		t.Error("Generated types.go missing Query method in DB interface")
	}
	if !strings.Contains(typesStr, "Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)") {
		t.Error("Generated types.go missing Exec method in DB interface")
	}

	// Check for CreateValidator interface
	if !strings.Contains(typesStr, "type CreateValidator interface") {
		t.Error("Generated types.go missing 'type CreateValidator interface'")
	}
	if !strings.Contains(typesStr, "ValidateCreate(ctx context.Context, input any) error") {
		t.Error("Generated types.go missing ValidateCreate method")
	}

	// Check for UpdateValidator interface
	if !strings.Contains(typesStr, "type UpdateValidator interface") {
		t.Error("Generated types.go missing 'type UpdateValidator interface'")
	}
	if !strings.Contains(typesStr, "ValidateUpdate(ctx context.Context, id any, input any) error") {
		t.Error("Generated types.go missing ValidateUpdate method")
	}

	// Check for Registry struct
	if !strings.Contains(typesStr, "type Registry struct") {
		t.Error("Generated types.go missing 'type Registry struct'")
	}
	if !strings.Contains(typesStr, "func NewRegistry() *Registry") {
		t.Error("Generated types.go missing NewRegistry function")
	}
	if !strings.Contains(typesStr, "func (r *Registry) Register(name string, action any)") {
		t.Error("Generated types.go missing Register method")
	}
	if !strings.Contains(typesStr, "func (r *Registry) Get(name string) (any, bool)") {
		t.Error("Generated types.go missing Get method")
	}
	if !strings.Contains(typesStr, "func GetTyped[T any](r *Registry, name string) (T, bool)") {
		t.Error("Generated types.go missing GetTyped generic function")
	}

	// Verify product.go was generated
	productPath := filepath.Join(tempDir, "actions", "product.go")
	productContent, err := os.ReadFile(productPath)
	if err != nil {
		t.Fatalf("Failed to read generated product.go: %v", err)
	}

	productStr := string(productContent)

	// Check for ProductActions interface
	if !strings.Contains(productStr, "type ProductActions interface") {
		t.Error("Generated product.go missing 'type ProductActions interface'")
	}

	// Check for all 5 CRUD methods in interface
	if !strings.Contains(productStr, "List(ctx context.Context, filter models.ProductFilter, sort models.ProductSort, page int, pageSize int) ([]models.Product, int64, error)") {
		t.Error("Generated product.go missing List method signature")
	}
	if !strings.Contains(productStr, "Get(ctx context.Context, id uuid.UUID) (*models.Product, error)") {
		t.Error("Generated product.go missing Get method signature")
	}
	if !strings.Contains(productStr, "Create(ctx context.Context, input models.ProductCreate) (*models.Product, error)") {
		t.Error("Generated product.go missing Create method signature")
	}
	if !strings.Contains(productStr, "Update(ctx context.Context, id uuid.UUID, input models.ProductUpdate) (*models.Product, error)") {
		t.Error("Generated product.go missing Update method signature")
	}
	if !strings.Contains(productStr, "Delete(ctx context.Context, id uuid.UUID) error") {
		t.Error("Generated product.go missing Delete method signature")
	}

	// Check for DefaultProductActions struct
	if !strings.Contains(productStr, "type DefaultProductActions struct") {
		t.Error("Generated product.go missing 'type DefaultProductActions struct'")
	}
	if !strings.Contains(productStr, "DB DB") {
		t.Error("Generated product.go missing DB field in DefaultProductActions")
	}

	// Check for method implementations
	if !strings.Contains(productStr, "func (a *DefaultProductActions) List(") {
		t.Error("Generated product.go missing List implementation")
	}
	if !strings.Contains(productStr, "func (a *DefaultProductActions) Get(") {
		t.Error("Generated product.go missing Get implementation")
	}
	if !strings.Contains(productStr, "func (a *DefaultProductActions) Create(") {
		t.Error("Generated product.go missing Create implementation")
	}
	if !strings.Contains(productStr, "func (a *DefaultProductActions) Update(") {
		t.Error("Generated product.go missing Update implementation")
	}
	if !strings.Contains(productStr, "func (a *DefaultProductActions) Delete(") {
		t.Error("Generated product.go missing Delete implementation")
	}

	// Check validation wiring in Create method
	if !strings.Contains(productStr, "validation.ValidateProductCreate(input)") {
		t.Error("Generated product.go Create method missing validation call")
	}
	if !strings.Contains(productStr, "errors.NewValidationError(valErrs)") {
		t.Error("Generated product.go Create method missing validation error wrapping")
	}

	// Check validation wiring in Update method
	if !strings.Contains(productStr, "validation.ValidateProductUpdate(input)") {
		t.Error("Generated product.go Update method missing validation call")
	}

	// Check error wiring
	if !strings.Contains(productStr, "errors.NotFound") {
		t.Error("Generated product.go missing errors.NotFound usage")
	}
	if !strings.Contains(productStr, "errors.MapDBError") {
		t.Error("Generated product.go missing errors.MapDBError usage")
	}

	// Check query integration
	if !strings.Contains(productStr, "queries.ProductFilterMods(filter)") {
		t.Error("Generated product.go List method missing filter mods call")
	}
	if !strings.Contains(productStr, "queries.ProductSortMod(sort)") {
		t.Error("Generated product.go List method missing sort mod call")
	}

	// Check real CRUD implementations (not stubs)
	if !strings.Contains(productStr, "pgx.CollectOneRow") {
		t.Error("Generated product.go missing pgx.CollectOneRow for Get/Create/Update")
	}
	if !strings.Contains(productStr, "pgx.CollectRows") {
		t.Error("Generated product.go missing pgx.CollectRows for List")
	}
	if !strings.Contains(productStr, "INSERT INTO products") {
		t.Error("Generated product.go Create missing INSERT INTO SQL")
	}
	if !strings.Contains(productStr, "UPDATE products SET") {
		t.Error("Generated product.go Update missing UPDATE SET SQL")
	}
	if !strings.Contains(productStr, "DELETE FROM products") {
		t.Error("Generated product.go Delete missing DELETE FROM SQL")
	}
	if !strings.Contains(productStr, "psql.Select") {
		t.Error("Generated product.go List missing Bob psql.Select")
	}
	if !strings.Contains(productStr, "RETURNING *") {
		t.Error("Generated product.go Create/Update missing RETURNING *")
	}
}

func TestGenerateActions_MultipleResources(t *testing.T) {
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

	// Call GenerateActions
	err := GenerateActions(resources, tempDir, "github.com/example/testapp")
	if err != nil {
		t.Fatalf("GenerateActions failed: %v", err)
	}

	// Verify types.go exists
	typesPath := filepath.Join(tempDir, "actions", "types.go")
	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		t.Error("types.go was not generated")
	}

	// Verify product.go exists
	productPath := filepath.Join(tempDir, "actions", "product.go")
	if _, err := os.Stat(productPath); os.IsNotExist(err) {
		t.Error("product.go was not generated")
	}

	// Verify category.go exists
	categoryPath := filepath.Join(tempDir, "actions", "category.go")
	if _, err := os.Stat(categoryPath); os.IsNotExist(err) {
		t.Error("category.go was not generated")
	}

	// Verify category.go contains CategoryActions interface
	categoryContent, err := os.ReadFile(categoryPath)
	if err != nil {
		t.Fatalf("Failed to read generated category.go: %v", err)
	}

	categoryStr := string(categoryContent)
	if !strings.Contains(categoryStr, "type CategoryActions interface") {
		t.Error("Generated category.go missing 'type CategoryActions interface'")
	}
	if !strings.Contains(categoryStr, "type DefaultCategoryActions struct") {
		t.Error("Generated category.go missing 'type DefaultCategoryActions struct'")
	}
}
