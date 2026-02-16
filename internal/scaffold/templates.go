package scaffold

import "embed"

// TemplatesFS contains embedded project scaffolding templates
//
//go:embed templates/*
var TemplatesFS embed.FS
