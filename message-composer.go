package main

import (
	"bytes"
	"html/template"
	"log"
	"os"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go"
)

type MessageTemplate struct {
	To      *template.Template
	Subject *template.Template
	HTML    *template.Template
	Text    *template.Template
}
type MessageComposer struct {
	Templates map[string]MessageTemplate
}

func NewMessageComposer() (t *MessageComposer, err error) {
	eventType := os.Getenv("EVENT_TYPE")

	templates := map[string]MessageTemplate{}

	templates[eventType] = MessageTemplate{
		To:      template.Must(template.New("TO_TEMPLATE").Parse(os.Getenv("TO_TEMPLATE"))),
		Subject: template.Must(template.New("SUBJECT_TEMPLATE").Parse(os.Getenv("SUBJECT_TEMPLATE"))),
		HTML:    template.Must(template.New("HTML_TEMPLATE").Parse(os.Getenv("HTML_TEMPLATE"))),
		Text:    template.Must(template.New("TEXT_TEMPLATE").Parse(os.Getenv("TEXT_TEMPLATE"))),
	}

	t = &MessageComposer{
		templates,
	}

	log.Println("created message composer with templates", t.templateKeys())

	return
}

func (m *MessageComposer) templateKeys() (keys []string) {
	keys = make([]string, 0, len(m.Templates))
	for k := range m.Templates {
		keys = append(keys, k)
	}
	return
}

func (t *MessageComposer) MessageFromEvent(event cloudevents.Event) (m *SMTPTransportMessage, err error) {
	template, ok := t.Templates[event.Type()]

	if !ok {
		return
	}

	data := event.Data

	to, err := ExecuteTemplate(template.To, data)
	if err != nil {
		return
	}
	subject, err := ExecuteTemplate(template.Subject, data)
	if err != nil {
		return
	}
	html, err := ExecuteTemplate(template.HTML, data)
	if err != nil {
		return
	}
	text, err := ExecuteTemplate(template.Text, data)
	if err != nil {
		return
	}
	m = &SMTPTransportMessage{
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
