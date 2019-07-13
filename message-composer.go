package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/ghodss/yaml"
)

type MessageTemplateSourceValue interface{}

func getTemplateFromSourceValue(v MessageTemplateSourceValue) *template.Template {
	tPath := os.Getenv("TEMPLATES_PATH")

	funcMap := template.FuncMap{
		"env": func(key string) string {
			return os.Getenv(key)
		},
	}

	switch i := v.(type) {
	case string:
		return template.Must(template.New("template").Funcs(funcMap).Parse(v.(string)))
	case map[string]interface{}:
		if i["file"] != "" {
			name := i["file"].(string)
			b, err := ioutil.ReadFile(path.Join(tPath, name))
			if err != nil {
				panic(err)
			}
			fmt.Println("??", string(b))
			return template.Must(template.New(name).Funcs(funcMap).Parse(string(b)))
		}
	}
	return nil
}

type MessageTemplateSource struct {
	To      MessageTemplateSourceValue `json:"to" yaml:"to"`
	Subject MessageTemplateSourceValue `json:"subject" yaml:"subject"`
	HTML    MessageTemplateSourceValue `json:"html" yaml:"html"`
	Text    MessageTemplateSourceValue `json:"text" yaml:"text"`
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
				if path.Ext(filename) != ".yaml" && path.Ext(filename) != ".yml" {
					continue
				}
				dat, err := ioutil.ReadFile(path.Join(tPath, filename))
				if err != nil {
					return t, err
				}

				var source MessageTemplateSource
				err = yaml.Unmarshal(dat, &source)
				if err != nil {
					return t, fmt.Errorf("Could not parse file %s, err: %s", filename, err.Error())
				}

				templates[strings.TrimSuffix(filename, path.Ext(filename))] = MessageTemplate{
					To:      getTemplateFromSourceValue(source.To),
					Subject: getTemplateFromSourceValue(source.Subject),
					HTML:    getTemplateFromSourceValue(source.HTML),
					Text:    getTemplateFromSourceValue(source.Text),
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
