package schema

import "testing"

func TestProductSchema(t *testing.T) {
	// Create the example Product schema from the plan
	Product := Define("Product",
		SoftDelete(),
		UUID("ID").PrimaryKey(),
		String("Title").Required().MaxLen(200).Label("Product Title"),
		Text("Description").Help("Full product description"),
		Decimal("Price").Required().Filterable().Sortable(),
		Enum("Status", "draft", "published", "archived").Default("draft"),
		Bool("Featured").Default(false),
		BelongsTo("Category", "categories").Optional().OnDelete(SetNull),
		HasMany("Reviews", "reviews"),
		Timestamps(),
	)

	// Verify the definition was populated correctly
	if Product.Name() != "Product" {
		t.Errorf("Expected name 'Product', got '%s'", Product.Name())
	}

	// Check field count (ID, Title, Description, Price, Status, Featured = 6 fields)
	if len(Product.Fields()) != 6 {
		t.Errorf("Expected 6 fields, got %d", len(Product.Fields()))
	}

	// Check relationship count (Category, Reviews = 2 relationships)
	if len(Product.Relationships()) != 2 {
		t.Errorf("Expected 2 relationships, got %d", len(Product.Relationships()))
	}

	// Check option count (SoftDelete = 1 option)
	if len(Product.Options()) != 1 {
		t.Errorf("Expected 1 option, got %d", len(Product.Options()))
	}

	// Check timestamps
	if !Product.HasTimestamps() {
		t.Error("Expected HasTimestamps to be true")
	}

	// Verify specific field types
	fields := Product.Fields()
	if fields[0].Type() != TypeUUID {
		t.Errorf("Expected first field to be UUID, got %s", fields[0].Type().String())
	}
	if fields[1].Type() != TypeString {
		t.Errorf("Expected second field to be String, got %s", fields[1].Type().String())
	}
	if fields[4].Type() != TypeEnum {
		t.Errorf("Expected fifth field to be Enum, got %s", fields[4].Type().String())
	}

	// Verify enum values
	enumValues := fields[4].EnumValues()
	if len(enumValues) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enumValues))
	}
	if enumValues[0] != "draft" || enumValues[1] != "published" || enumValues[2] != "archived" {
		t.Errorf("Unexpected enum values: %v", enumValues)
	}

	// Verify relationships
	rels := Product.Relationships()
	if rels[0].RelType() != RelBelongsTo {
		t.Errorf("Expected first relationship to be BelongsTo, got %s", rels[0].RelType().String())
	}
	if rels[0].Table() != "categories" {
		t.Errorf("Expected first relationship table to be 'categories', got '%s'", rels[0].Table())
	}
	if !rels[0].IsOptional() {
		t.Error("Expected first relationship to be optional")
	}
	if rels[0].OnDeleteAction() != SetNull {
		t.Errorf("Expected first relationship OnDelete to be SetNull, got %s", rels[0].OnDeleteAction().String())
	}

	if rels[1].RelType() != RelHasMany {
		t.Errorf("Expected second relationship to be HasMany, got %s", rels[1].RelType().String())
	}

	// Verify option
	opts := Product.Options()
	if opts[0].Type() != OptSoftDelete {
		t.Errorf("Expected option to be SoftDelete, got %s", opts[0].Type().String())
	}
}

func TestAllFieldTypes(t *testing.T) {
	// Verify all 14 field type constructors work
	def := Define("TestResource",
		UUID("field1"),
		String("field2"),
		Text("field3"),
		Int("field4"),
		BigInt("field5"),
		Decimal("field6"),
		Bool("field7"),
		DateTime("field8"),
		Date("field9"),
		Enum("field10", "a", "b", "c"),
		JSON("field11"),
		Slug("field12"),
		Email("field13"),
		URL("field14"),
	)

	if len(def.Fields()) != 14 {
		t.Errorf("Expected 14 fields, got %d", len(def.Fields()))
	}

	expectedTypes := []FieldType{
		TypeUUID, TypeString, TypeText, TypeInt, TypeBigInt, TypeDecimal,
		TypeBool, TypeDateTime, TypeDate, TypeEnum, TypeJSON, TypeSlug,
		TypeEmail, TypeURL,
	}

	for i, field := range def.Fields() {
		if field.Type() != expectedTypes[i] {
			t.Errorf("Field %d: expected type %s, got %s",
				i, expectedTypes[i].String(), field.Type().String())
		}
	}
}

func TestAllRelationshipTypes(t *testing.T) {
	// Verify all 4 relationship types work
	def := Define("TestResource",
		BelongsTo("rel1", "table1"),
		HasMany("rel2", "table2"),
		HasOne("rel3", "table3"),
		ManyToMany("rel4", "table4"),
	)

	if len(def.Relationships()) != 4 {
		t.Errorf("Expected 4 relationships, got %d", len(def.Relationships()))
	}

	expectedTypes := []RelationType{RelBelongsTo, RelHasMany, RelHasOne, RelManyToMany}

	for i, rel := range def.Relationships() {
		if rel.RelType() != expectedTypes[i] {
			t.Errorf("Relationship %d: expected type %s, got %s",
				i, expectedTypes[i].String(), rel.RelType().String())
		}
	}
}

func TestAllOptions(t *testing.T) {
	// Verify all 4 options work
	def := Define("TestResource",
		SoftDelete(),
		Auditable(),
		TenantScoped(),
		Searchable(),
	)

	if len(def.Options()) != 4 {
		t.Errorf("Expected 4 options, got %d", len(def.Options()))
	}

	expectedTypes := []OptionType{OptSoftDelete, OptAuditable, OptTenantScoped, OptSearchable}

	for i, opt := range def.Options() {
		if opt.Type() != expectedTypes[i] {
			t.Errorf("Option %d: expected type %s, got %s",
				i, expectedTypes[i].String(), opt.Type().String())
		}
	}
}

func TestFluentChaining(t *testing.T) {
	// Verify fluent chaining returns correct type for continued chaining
	field := String("Test").Required().MaxLen(100).MinLen(5).Unique().Index().Sortable().Filterable()

	if field.Name() != "Test" {
		t.Errorf("Expected name 'Test', got '%s'", field.Name())
	}

	// Count modifiers
	if len(field.Modifiers()) != 7 {
		t.Errorf("Expected 7 modifiers, got %d", len(field.Modifiers()))
	}
}
