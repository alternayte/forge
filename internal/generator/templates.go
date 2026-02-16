package generator

import "embed"

// TemplatesFS contains embedded code generation templates
//
//go:embed templates/*
var TemplatesFS embed.FS
