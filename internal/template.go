package internal

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"log"
	"time"
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
		tpl, err := template.New(t.Name).Funcs(template.FuncMap{
			"checksum": func(content string) string {
				return fmt.Sprintf("%x", md5.Sum([]byte(content)))
			},
			"summary": func(content string, skip, length int) string {
				runes := []rune(content)
				if len(runes) > (skip + length) {
					return string(runes[skip:skip+length-1]) + " ..."
				}
				return string(runes[skip:])
			},
			"add": func(a, b int64) int64 {
				return a + b
			},
			"utc2milli": func(utc string) int64 {
				utcTime, err := time.Parse("2006-01-02T15:04:05Z", utc)
				if err != nil {
					return -1
				}
				return utcTime.UnixMilli()
			},
		}).Parse(t.Content)
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
