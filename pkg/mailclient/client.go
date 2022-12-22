package mailclient

import (
	"context"
	"io"
)

// Client is an interface to send client
type Client interface {
	io.Closer
	SendEmails(ctx context.Context, parsedEmails []EmailSingle) (report Report)
}

// EmailSingle is parsed and ready to send email.
type EmailSingle struct {
	TrackingID string `json:"tracking_id" validate:"required"`
	SenderAddr string `json:"sender_addr" validate:"required"`

	// Recipients is only used as the list of email recipient, but we'll send it as single email using To address.
	Recipients  []string          `json:"recipients" validate:"required"`
	Subject     string            `json:"subject" validate:"required"`
	Body        string            `json:"body" validate:"required"`
	Attachments map[string]string `json:"attachments" validate:"-"`
}

type RecvReport struct {
	To        string
	Error     error
	EmailData EmailSingle
}

type Report struct {
	ClientError error
	RecvReports []RecvReport
}
