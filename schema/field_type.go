package schema

// UUID creates a UUID field with the given name.
func UUID(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeUUID,
	}
}

// String creates a string field with the given name.
func String(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeString,
	}
}

// Text creates a text field with the given name.
func Text(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeText,
	}
}

// Int creates an integer field with the given name.
func Int(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeInt,
	}
}

// BigInt creates a big integer field with the given name.
func BigInt(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeBigInt,
	}
}

// Decimal creates a decimal field with the given name.
func Decimal(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeDecimal,
	}
}

// Bool creates a boolean field with the given name.
func Bool(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeBool,
	}
}

// DateTime creates a date-time field with the given name.
func DateTime(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeDateTime,
	}
}

// Date creates a date field with the given name.
func Date(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeDate,
	}
}

// Enum creates an enum field with the given name and allowed values.
func Enum(name string, values ...string) *Field {
	return &Field{
		name:       name,
		fieldType:  TypeEnum,
		enumValues: values,
	}
}

// JSON creates a JSON field with the given name.
func JSON(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeJSON,
	}
}

// Slug creates a slug field with the given name.
func Slug(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeSlug,
	}
}

// Email creates an email field with the given name.
func Email(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeEmail,
	}
}

// URL creates a URL field with the given name.
func URL(name string) *Field {
	return &Field{
		name:      name,
		fieldType: TypeURL,
	}
}
