package utils

import (
	"gopkg.in/gomail.v2"
	"uwece.ca/app/config"
)

type Mailer struct {
	mailer *gomail.Dialer
	cfg    *config.Config
}

func NewMailer(cfg *config.Config) *Mailer {
	m := cfg.Mailer
	d := gomail.NewDialer(m.Host, m.Port, m.Username, m.Password)

	return &Mailer{
		mailer: d,
		cfg:    cfg,
	}
}

type Email struct {
	To       string
	Name     string
	Subject  string
	TextBody string
	HtmlBody string
}

func (m *Mailer) SendMessage(me Email) error {
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
