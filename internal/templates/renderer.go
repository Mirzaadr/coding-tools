// Package templates provides the low-level text/template rendering engine
// used to produce generated source files from .tmpl files stored under the
// top-level templates/ directory. It knows nothing about builders,
// architectures, or the CLI - it only renders data into text.
package templates

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

// Engine renders .tmpl files with a given data context.
type Engine struct {
	// RootDir is the base directory that template paths are resolved
	// relative to (e.g. "templates/clean").
	RootDir string
}

// NewEngine constructs a template Engine rooted at rootDir.
func NewEngine(rootDir string) *Engine {
	return &Engine{RootDir: rootDir}
}

// Render loads the template at templateRelativePath (relative to RootDir),
// executes it against data, and returns the rendered bytes.
func (e *Engine) Render(templateRelativePath string, data any) ([]byte, error) {
	fullPath := templateRelativePath
	if e.RootDir != "" {
		fullPath = e.RootDir + string(os.PathSeparator) + templateRelativePath
	}

	tmplContent, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("templates: failed to read template %q: %w", fullPath, err)
	}

	tmpl, err := template.New(templateRelativePath).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("templates: failed to parse template %q: %w", fullPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("templates: failed to execute template %q: %w", fullPath, err)
	}

	return buf.Bytes(), nil
}
