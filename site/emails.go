package site

import (
	"fmt"
	"html/template"

	"uwece.ca/app/auth"
	"uwece.ca/app/email"
)

func (s *Site) SendVerificationEmail(addr, name string, token auth.Token) error {
	params := struct {
		Name   string
		Link   template.HTML
		Scheme string
	}{
		Name:   name,
		Scheme: "https",
		Link:   template.HTML(fmt.Sprintf("%s/signup/verify/%s", s.config.Core.BaseDomain, token)),
	}

	if s.config.Core.Development {
		params.Scheme = "http"
	}

	text, err := s.templates.ExecutePlainString("emails/verification.txt", params)
	if err != nil {
		return fmt.Errorf("error rendering email text template: %w", err)
	}

	html, err := s.templates.ExecutePlainString("emails/verification", params)
	if err != nil {
		return fmt.Errorf("error rendering email html template: %w", err)
	}

	err = s.mailer.SendMessage(email.Message{
		To:       addr,
		Name:     name,
		Subject:  "UWECECA - Verification",
		TextBody: text,
		HtmlBody: html,
	})
	if err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}

	return nil
}
