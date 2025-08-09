package providers

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
)

type GmailSmtpProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewGmailSmtpProvider(host, port, username, password, from string) *GmailSmtpProvider {
	return &GmailSmtpProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (g *GmailSmtpProvider) SendEmail(ctx context.Context, to, subject, html string) error {
	addr := fmt.Sprintf("%s:%s", g.host, g.port)

	auth := smtp.PlainAuth("", g.username, g.password, g.host)

	msg := "From: " + g.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n" +
		html

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         g.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		log.Printf("SMTP dial error: %v", err)
		return err
	}

	c, err := smtp.NewClient(conn, g.host)
	if err != nil {
		log.Printf("SMTP client error: %v", err)
		return err
	}

	if err = c.Auth(auth); err != nil {
		log.Printf("SMTP auth error: %v", err)
		return err
	}

	if err = c.Mail(g.from); err != nil {
		return err
	}
	if err = c.Rcpt(to); err != nil {
		return err
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	log.Printf("Gmail SMTP email sent to %s", to)
	return nil
}