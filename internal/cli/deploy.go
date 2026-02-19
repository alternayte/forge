package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/forge-framework/forge/internal/config"
	"github.com/forge-framework/forge/internal/ui"
	"github.com/spf13/cobra"
)

// dockerfileTemplate is the multi-stage Dockerfile for production deployment.
// Uses alpine base with HEALTHCHECK on admin port 9090.
const dockerfileTemplate = `# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -ldflags="-s -w \
    -X main.Version=$(git describe --tags --always 2>/dev/null || echo dev) \
    -X main.Commit=$(git rev-parse --short HEAD 2>/dev/null || echo none) \
    -X main.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /app/bin/{{ .ProjectName }} .

# Runtime stage
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/bin/{{ .ProjectName }} /usr/local/bin/{{ .ProjectName }}
EXPOSE 8080 9090
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:9090/healthz || exit 1
CMD ["{{ .ProjectName }}", "serve"]
`

// deployTemplateData holds the values injected into dockerfileTemplate.
type deployTemplateData struct {
	ProjectName string
}

func newDeployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Generate Dockerfile for deployment",
		Long:  `Generates a production Dockerfile in the project root. The Dockerfile uses a multi-stage build to produce a minimal container image.`,
		RunE:  runDeploy,
	}
}

func runDeploy(cmd *cobra.Command, args []string) error {
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

	// Determine project name (fallback to "app")
	projectName := cfg.Project.Name
	if projectName == "" {
		projectName = "app"
	}

	// Check if Dockerfile already exists
	dockerfilePath := filepath.Join(projectRoot, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		fmt.Fprintf(os.Stderr, "%s Dockerfile already exists, overwriting\n", ui.WarnIcon)
	}

	// Render Dockerfile from template
	tmpl, err := template.New("dockerfile").Parse(dockerfileTemplate)
	if err != nil {
		return fmt.Errorf("parsing Dockerfile template: %w", err)
	}

	f, err := os.Create(dockerfilePath)
	if err != nil {
		return fmt.Errorf("creating Dockerfile: %w", err)
	}
	defer f.Close()

	data := deployTemplateData{
		ProjectName: projectName,
	}
	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("rendering Dockerfile: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.Success(fmt.Sprintf("Dockerfile generated at %s", dockerfilePath)))
	fmt.Println()
	fmt.Println(ui.Info("Next steps:"))
	fmt.Printf("  docker build -t %s .\n", projectName)
	fmt.Printf("  docker run -p 8080:8080 -p 9090:9090 %s\n", projectName)
	fmt.Println()

	return nil
}
