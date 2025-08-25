package mailer

import (
	"embed"
	"fmt"

	"gopkg.in/gomail.v2"
	"uwece.ca/app/config"
	"uwece.ca/app/templates"
	"uwece.ca/app/utils"
)

//go:embed templates/*
var embedFS embed.FS

type Mailer interface {
	SendVerificationEmail(addr, name string, token utils.Token) error
}

type mailer struct {
	cfg       *config.Config
	mailer    *gomail.Dialer
	templates *templates.Templates
}

func New(cfg *config.Config) Mailer {
	m := cfg.Mailer
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)

	var tmpl *templates.Templates
	if cfg.Core.Development {
		tmpl = templates.NewDevTemplates(embedFS, "./mailer/templates")
	} else {
		tmpl = templates.NewTemplates(embedFS)
	}

	return &mailer{
		cfg:       cfg,
		mailer:    d,
		templates: tmpl,
	}
}

func (m *mailer) SendVerificationEmail(addr, name string, token utils.Token) error {
	scheme := "https"
	if m.cfg.Core.Development {
		scheme = "http"
	}

	params := templates.Context{
		"name": name,
		"link": fmt.Sprintf("%s://%s/signup/verify/%s", scheme, m.cfg.Core.BaseDomain, token),
	}

	text, err := m.templates.ExecutePlainString("verification.txt", params)
	if err != nil {
		return fmt.Errorf("error rendering email text template: %w", err)
	}

	html, err := m.templates.ExecutePlainString("verification", params)
	if err != nil {
		return fmt.Errorf("error rendering email html template: %w", err)
	}

	err = m.sendMessage(email{
		To:       addr,
		Name:     addr,
		Subject:  "UWECECA - Verification",
		TextBody: text,
		HtmlBody: html,
	})
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}

type email struct {
	To       string
	Name     string
	Subject  string
	TextBody string
	HtmlBody string
}

func (m *mailer) sendMessage(me email) error {
	message := gomail.NewMessage()

	message.SetAddressHeader("From", m.cfg.Mailer.FromAddress, "UWECECA Team")
	message.SetHeader("Subject", me.Subject)
	message.SetAddressHeader("To", me.To, me.Name)
	message.SetBody("text/plain", me.TextBody)
	if me.HtmlBody != "" {
		message.AddAlternative("text/html", me.HtmlBody)
	}

	return m.mailer.DialAndSend(message)
}
