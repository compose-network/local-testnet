package blockscout

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed nginx.conf.tmpl
var nginxTemplateFS embed.FS

const nginxTemplate = "nginx.conf.tmpl"

type nginxTemplateData struct {
	ServiceSuffix string
}

func GenerateNginxConfigs(networksDir string) error {
	tmplContent, err := nginxTemplateFS.ReadFile(nginxTemplate)
	if err != nil {
		return fmt.Errorf("failed to read embedded %s: %w", nginxTemplate, err)
	}

	tmpl, err := template.New("nginx").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse nginx template: %w", err)
	}

	rollups := []struct {
		name   string
		suffix string
	}{
		{"rollup-a", "a"},
		{"rollup-b", "b"},
	}

	for _, rollup := range rollups {
		rollupDir := filepath.Join(networksDir, rollup.name)
		if err := os.MkdirAll(rollupDir, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %w", rollupDir, err)
		}

		var buf bytes.Buffer
		data := nginxTemplateData{ServiceSuffix: rollup.suffix}
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", rollup.name, err)
		}

		nginxConfPath := filepath.Join(rollupDir, "blockscout-nginx.conf")
		if err := os.WriteFile(nginxConfPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", nginxConfPath, err)
		}
	}

	return nil
}
