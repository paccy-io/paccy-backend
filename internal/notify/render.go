package notify

import (
	"bytes"
	"html/template"
	"path/filepath"
)

func RenderTemplate(name string, data any) (string, error) {
	path := filepath.Join("internal", "notify", "templates", name)

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// is valid template name
func IsValidTemplateName(name string) bool {
	// check if the file exists under the templates directory
	path := filepath.Join("internal", "notify", "templates", name)
	_, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	return true
}
