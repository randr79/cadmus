package build

import (
	"bytes"
	"fmt"
	"text/template"
)

type Main struct {
	mainTemplate *template.Template
	imports      map[string]string // Wordt nu niet gebruikt in de template, maar handig voor later
	entrypoints  map[string]string // Key: "net use", Value: "NET_USE_Adapter"
}

func (m Main) Render() (string, error) {
	// 1. Maak een buffer aan om de gegenereerde code in op te vangen
	var buf bytes.Buffer

	// 2. Bereid de data voor die de template verwacht
	templateData := struct {
		Entrypoints map[string]string
	}{
		Entrypoints: m.entrypoints,
	}

	// 3. Voer het template uit en geef de error netjes terug
	if err := m.mainTemplate.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute main template: %w", err)
	}

	// 4. Geef de gegenereerde Go-code terug als string
	return buf.String(), nil
}
