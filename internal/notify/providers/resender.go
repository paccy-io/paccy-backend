package providers

import (
	"context"
	"fmt"
	"log"

	"github.com/resend/resend-go/v2"
)

type ResendProvider struct {
	client   *resend.Client
	fromAddr string
}

func NewResendProvider(apiKey, from string) *ResendProvider {
	client := resend.NewClient(apiKey)
	return &ResendProvider{
		client:   client,
		fromAddr: from,
	}
}

func (r *ResendProvider) SendEmail(ctx context.Context, to, subject, html string) error {
	input := &resend.SendEmailRequest{
		From:    r.fromAddr,
		To:      []string{to},
		Subject: subject,
		Html:    html,
		ReplyTo: "support@paccy.io", //!! will be fixed later
	}

	_, err := r.client.Emails.Send(input)
	if err != nil {
		log.Printf("Resend send failed: %v", err)
		return fmt.Errorf("resend provider: %w", err)
	}

	log.Printf("Resend email sent to %s", to)
	return nil
}
