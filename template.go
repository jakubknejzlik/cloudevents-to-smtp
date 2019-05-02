package main

import (
	"bytes"
	"html/template"
	"os"
	"strings"
)

type TemplateHandler struct {
	ToTemplate      *template.Template
	SubjectTemplate *template.Template
	HTMLTemplate    *template.Template
	TextTemplate    *template.Template
}

func NewTemplateHandler() (t *TemplateHandler, err error) {
	t = &TemplateHandler{
		ToTemplate:      template.Must(template.New("TO_TEMPLATE").Parse(os.Getenv("TO_TEMPLATE"))),
		SubjectTemplate: template.Must(template.New("SUBJECT_TEMPLATE").Parse(os.Getenv("SUBJECT_TEMPLATE"))),
		HTMLTemplate:    template.Must(template.New("HTML_TEMPLATE").Parse(os.Getenv("HTML_TEMPLATE"))),
		TextTemplate:    template.Must(template.New("TEXT_TEMPLATE").Parse(os.Getenv("TEXT_TEMPLATE"))),
	}
	return
}

func (t *TemplateHandler) MessageFromData(data interface{}) (m SMTPTransportMessage, err error) {
	to, err := ExecuteTemplate(t.ToTemplate, data)
	if err != nil {
		return
	}
	subject, err := ExecuteTemplate(t.SubjectTemplate, data)
	if err != nil {
		return
	}
	html, err := ExecuteTemplate(t.HTMLTemplate, data)
	if err != nil {
		return
	}
	text, err := ExecuteTemplate(t.TextTemplate, data)
	if err != nil {
		return
	}
	m = SMTPTransportMessage{
		To:      strings.Split(to, ","),
		Subject: subject,
		HTML:    html,
		Text:    text,
	}
	return
}

func ExecuteTemplate(temp *template.Template, data interface{}) (string, error) {
	b := new(bytes.Buffer)
	err := temp.Execute(b, data)
	return b.String(), err
}
