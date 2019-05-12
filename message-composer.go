package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/ghodss/yaml"
)

type MessageTemplateSource struct {
	To      string `json:"to" yaml:"to"`
	Subject string `json:"subject" yaml:"subject"`
	HTML    string `json:"html" yaml:"html"`
	Text    string `json:"text" yaml:"text"`
}
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

	funcMap := template.FuncMap{
		"env": func(key string) string {
			return os.Getenv(key)
		},
	}

	templates[eventType] = MessageTemplate{
		To:      template.Must(template.New("TO_TEMPLATE").Funcs(funcMap).Parse(os.Getenv("TO_TEMPLATE"))),
		Subject: template.Must(template.New("SUBJECT_TEMPLATE").Funcs(funcMap).Parse(os.Getenv("SUBJECT_TEMPLATE"))),
		HTML:    template.Must(template.New("HTML_TEMPLATE").Funcs(funcMap).Parse(os.Getenv("HTML_TEMPLATE"))),
		Text:    template.Must(template.New("TEXT_TEMPLATE").Funcs(funcMap).Parse(os.Getenv("TEXT_TEMPLATE"))),
	}

	tPath := os.Getenv("TEMPLATES_PATH")
	if tPath != "" {
		files, err := ioutil.ReadDir(tPath)
		if err != nil {
			return t, err
		}
		for _, f := range files {
			if !f.IsDir() {
				filename := f.Name()
				dat, err := ioutil.ReadFile(path.Join(tPath, f.Name()))
				if err != nil {
					return t, err
				}

				var source MessageTemplateSource
				err = yaml.Unmarshal(dat, &source)
				if err != nil {
					return t, err
				}

				templates[strings.TrimSuffix(filename, path.Ext(filename))] = MessageTemplate{
					To:      template.Must(template.New("TO_TEMPLATE").Funcs(funcMap).Parse(source.To)),
					Subject: template.Must(template.New("SUBJECT_TEMPLATE").Funcs(funcMap).Parse(source.Subject)),
					HTML:    template.Must(template.New("HTML_TEMPLATE").Funcs(funcMap).Parse(source.HTML)),
					Text:    template.Must(template.New("TEXT_TEMPLATE").Funcs(funcMap).Parse(source.Text)),
				}
			}
			// return t, err
		}
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

	data := &map[string]interface{}{}
	if err := event.DataAs(data); err != nil {
		return m, err
	}

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
