package main

import (
	"encoding/gob"
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	gomail "gopkg.in/gomail.v2"
)

// SMTPTransport ...
type SMTPTransport struct {
	dialer *gomail.Dialer
	sender string
}

// SMTPTransportMessage ...
type SMTPTransportMessage struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

// NewSMTPTransport ...
func NewSMTPTransport(urlString, sender string) *SMTPTransport {

	gob.Register(SMTPTransportMessage{})
	URL, _ := url.Parse(urlString)

	if URL == nil {
		panic("SMTP url not provided")
	}
	if URL.User == nil {
		panic("user credentials not provided")
	}

	host := strings.Split(URL.Host, ":")[0]
	username := URL.User.Username()
	password := ""
	if pass, exists := URL.User.Password(); exists == true {
		password = pass
	}
	port := 25
	if portValue, err := strconv.ParseInt(URL.Port(), 10, 32); err == nil {
		port = int(portValue)
	}

	d := gomail.NewDialer(host, port, username, password)

	transport := SMTPTransport{dialer: d}
	transport.sender = sender
	return &transport
}

// SendMessage ...
func (t SMTPTransport) SendMessage(msg SMTPTransportMessage) error {
	sender := msg.From
	if sender == "" {
		sender = t.sender
	}

	address, err := mail.ParseAddress(sender)
	if err != nil {
		return err
	}

	fmt.Println("from", address.Address, address.Name)
	fmt.Println("to", msg.To)
	m := gomail.NewMessage()
	m.SetAddressHeader("From", address.Address, address.Name)
	m.SetHeader("To", msg.To...)
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/plain", msg.Text)
	m.SetBody("text/html", msg.HTML)
	// m.Attach("/home/Alex/lolcat.jpg")
	return t.dialer.DialAndSend(m)
}
