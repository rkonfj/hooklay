package internal

import (
	"bytes"
	"errors"
	"log"
	"text/template"
)

type Template struct {
	Name    string
	Content string
}

type TemplateManager struct {
	templates map[string]*template.Template
}

func NewTemplateManager(templates []*Template) *TemplateManager {
	templateMap := make(map[string]*template.Template)
	for _, t := range templates {
		log.Println("Register target body template " + t.Name)
		tpl, err := template.New(t.Name).Parse(t.Content)
		if err != nil {
			log.Fatal(err)
		}
		templateMap[t.Name] = tpl
	}
	return &TemplateManager{
		templates: templateMap,
	}
}

func (m *TemplateManager) Render(templateName string, rawRequestBody map[string]any) *bytes.Buffer {
	tpl := m.templates[templateName]
	if tpl == nil {
		log.Fatal(errors.New("no such template"))
	}
	var bodyBuffer bytes.Buffer
	err := tpl.Execute(&bodyBuffer, rawRequestBody)
	if err != nil {
		log.Fatal(err)
	}
	return &bodyBuffer
}
