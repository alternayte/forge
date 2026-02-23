package toolsync

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"time"
	"os"
	"path/filepath"
	"text/template"
)

var downloadClient = &http.Client{
	Timeout: 5 * time.Minute,
}

// DownloadTool downloads a tool binary from the registry.
func DownloadTool(tool ToolDef, platform Platform, destDir string, progress func(pct float64)) error {
	return DownloadToolWithContext(context.Background(), tool, platform, destDir, progress)
}

// DownloadToolWithContext downloads a tool binary from the registry with context support.
func DownloadToolWithContext(ctx context.Context, tool ToolDef, platform Platform, destDir string, progress func(pct float64)) error {
	// Apply OS/Arch mappings for tools that use non-standard platform names
	mappedPlatform := platform
	if mapped, ok := tool.OSMap[mappedPlatform.OS]; ok {
		mappedPlatform.OS = mapped
	}
	if mapped, ok := tool.ArchMap[mappedPlatform.Arch]; ok {
		mappedPlatform.Arch = mapped
	}

	// Construct URL from template
	url, err := constructURL(tool.URLTemplate, mappedPlatform, tool.Version)
	if err != nil {
		return fmt.Errorf("construct URL: %w", err)
	}

	// Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	// Download to memory buffer for checksum verification
	var buf bytes.Buffer
	if err := downloadWithProgress(ctx, url, &buf, progress); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// Verify checksum if provided
	if checksum, ok := tool.Checksums[platform.String()]; ok && checksum != "" {
		if err := verifyChecksum(buf.Bytes(), checksum); err != nil {
			return fmt.Errorf("checksum verification: %w", err)
		}
	}

	// Extract or write binary
	if tool.IsArchive {
		if err := extractTarGz(&buf, destDir, tool.BinaryName); err != nil {
			return fmt.Errorf("extract archive: %w", err)
		}
	} else {
		binPath := filepath.Join(destDir, tool.BinaryName)
		if err := os.WriteFile(binPath, buf.Bytes(), 0755); err != nil {
			return fmt.Errorf("write binary: %w", err)
		}
	}

	// Ensure binary is executable (Unix)
	binPath := filepath.Join(destDir, tool.BinaryName)
	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("chmod binary: %w", err)
	}

	return nil
}

// SyncAll downloads all tools, skipping those already installed unless forced.
func SyncAll(tools []ToolDef, platform Platform, destDir string, onProgress func(tool string, pct float64)) error {
	return SyncAllWithContext(context.Background(), tools, platform, destDir, onProgress)
}

// SyncAllWithContext downloads all tools with context support, skipping those already installed unless forced.
func SyncAllWithContext(ctx context.Context, tools []ToolDef, platform Platform, destDir string, onProgress func(tool string, pct float64)) error {
	for _, tool := range tools {
		if IsToolInstalled(destDir, tool.BinaryName) {
			// Already installed, skip
			if onProgress != nil {
				onProgress(tool.Name, 100.0)
			}
			continue
		}

		// Download tool
		if err := DownloadToolWithContext(ctx, tool, platform, destDir, func(pct float64) {
			if onProgress != nil {
				onProgress(tool.Name, pct)
			}
		}); err != nil {
			return fmt.Errorf("download %s: %w", tool.Name, err)
		}
	}
	return nil
}

// ToolBinPath returns the expected path to a tool binary.
func ToolBinPath(destDir, toolName string) string {
	return filepath.Join(destDir, toolName)
}

// IsToolInstalled checks if a tool binary exists at the expected path.
func IsToolInstalled(destDir, toolName string) bool {
	binPath := ToolBinPath(destDir, toolName)
	_, err := os.Stat(binPath)
	return err == nil
}

// constructURL renders the URL template with platform and version data.
func constructURL(urlTemplate string, platform Platform, version string) (string, error) {
	tmpl, err := template.New("url").Parse(urlTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		OS      string
		Arch    string
		Version string
	}{
		OS:      platform.OS,
		Arch:    platform.Arch,
		Version: version,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// downloadWithProgress downloads a URL to a writer, calling progress callback with percentage.
func downloadWithProgress(ctx context.Context, url string, w io.Writer, progress func(pct float64)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	total := resp.ContentLength
	var written int64

	// Create a TeeReader to track progress
	reader := io.TeeReader(resp.Body, &progressWriter{
		total:    total,
		written:  &written,
		callback: progress,
	})

	// Copy to destination
	if _, err := io.Copy(w, reader); err != nil {
		return err
	}

	// Report 100% completion
	if progress != nil {
		progress(100.0)
	}

	return nil
}

// progressWriter tracks bytes written and reports progress.
type progressWriter struct {
	total    int64
	written  *int64
	callback func(pct float64)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	*pw.written += int64(n)

	if pw.callback != nil && pw.total > 0 {
		pct := float64(*pw.written) / float64(pw.total) * 100.0
		pw.callback(pct)
	}

	return n, nil
}

// verifyChecksum checks if the data matches the expected SHA256 checksum.
func verifyChecksum(data []byte, expectedHex string) error {
	hash := sha256.Sum256(data)
	actualHex := hex.EncodeToString(hash[:])

	if actualHex != expectedHex {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, actualHex)
	}

	return nil
}

// extractTarGz extracts a tar.gz archive to destDir, extracting only the specified binary.
func extractTarGz(r io.Reader, destDir, binaryName string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		// Look for the binary file (may be at root or in a subdirectory)
		if filepath.Base(header.Name) == binaryName {
			destPath := filepath.Join(destDir, binaryName)

			// Create the binary file
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}

			outFile.Close()
			return nil // Binary found and extracted
		}
	}

	return fmt.Errorf("binary %s not found in archive", binaryName)
}
