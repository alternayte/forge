package errors

// ErrorCode is a unique identifier for each error type.
type ErrorCode string

// Parser errors (E0xx range)
const (
	ErrDynamicValue        ErrorCode = "E001" // Schema uses variable instead of literal
	ErrUnsupportedType     ErrorCode = "E002" // Field type not supported
	ErrInvalidFieldName    ErrorCode = "E003" // Field name invalid
	ErrMissingArgument     ErrorCode = "E004" // Required argument missing
	ErrInvalidModifierValue ErrorCode = "E005" // Modifier value invalid
)

// Schema validation errors (E1xx range)
const (
	ErrDuplicateField     ErrorCode = "E100" // Field name appears twice
	ErrInvalidEnumDefault ErrorCode = "E101" // Enum default not in allowed values
	ErrMissingPrimaryKey  ErrorCode = "E102" // Resource missing primary key
)

// ErrorInfo holds metadata about an error code.
type ErrorInfo struct {
	Code    ErrorCode // Error code
	Title   string    // Short description
	HelpURL string    // Help command or URL
}

// Registry maps error codes to their metadata.
var Registry = map[ErrorCode]ErrorInfo{
	ErrDynamicValue: {
		Code:    ErrDynamicValue,
		Title:   "schema values must use literal values for static analysis",
		HelpURL: "forge help E001",
	},
	ErrUnsupportedType: {
		Code:    ErrUnsupportedType,
		Title:   "unsupported field type",
		HelpURL: "forge help E002",
	},
	ErrInvalidFieldName: {
		Code:    ErrInvalidFieldName,
		Title:   "invalid field name",
		HelpURL: "forge help E003",
	},
	ErrMissingArgument: {
		Code:    ErrMissingArgument,
		Title:   "required argument missing",
		HelpURL: "forge help E004",
	},
	ErrInvalidModifierValue: {
		Code:    ErrInvalidModifierValue,
		Title:   "invalid modifier value",
		HelpURL: "forge help E005",
	},
	ErrDuplicateField: {
		Code:    ErrDuplicateField,
		Title:   "duplicate field name",
		HelpURL: "forge help E100",
	},
	ErrInvalidEnumDefault: {
		Code:    ErrInvalidEnumDefault,
		Title:   "enum default value not in allowed values",
		HelpURL: "forge help E101",
	},
	ErrMissingPrimaryKey: {
		Code:    ErrMissingPrimaryKey,
		Title:   "resource missing primary key",
		HelpURL: "forge help E102",
	},
}

// Lookup returns the ErrorInfo for a given code.
// Returns an empty ErrorInfo if the code is not registered.
func Lookup(code ErrorCode) ErrorInfo {
	info, ok := Registry[code]
	if !ok {
		return ErrorInfo{
			Code:    code,
			Title:   "unknown error",
			HelpURL: "forge help errors",
		}
	}
	return info
}
