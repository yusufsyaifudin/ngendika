package mailclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/go-playground/validator/v10"
	"go.uber.org/multierr"
	"io"
	"net"
	"strings"
	"sync"
)

type SmtpMailerConfig struct {
	EmailCredential *EmailCredential `validate:"required"`
}

type SmtpMailer struct {
	Config *SmtpMailerConfig
	smtp   *smtp.Client
	lock   sync.RWMutex
}

var _ Client = (*SmtpMailer)(nil)

// NewSmtp will return new smtp client without any real connection is made.
// It will connect on the first Send or SendEmails (except if it runs in dry-run mode).
func NewSmtp(cfg *SmtpMailerConfig) (*SmtpMailer, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		err = fmt.Errorf("validation error: %w", err)
		return nil, err
	}

	client := &SmtpMailer{
		Config: cfg,
	}

	return client, nil
}

func (m *SmtpMailer) SendEmails(ctx context.Context, parsedEmails []EmailSingle) (report Report) {
	recvReports := make([]RecvReport, 0)
	for _, emailData := range parsedEmails {
		// we need a report for each recipient, therefore we use single email send for single address.
		// Actually we can use LMTP, but not many email providers support LMTP, so we do a "trick" to send it one by one
		for _, to := range emailData.Recipients {
			recvReports = append(recvReports, m.sendEmail(ctx, to, emailData))
		}

	}

	report = Report{
		RecvReports: recvReports,
	}
	return
}

// SendEmail will do the real send email.
func (m *SmtpMailer) sendEmail(ctx context.Context, recvAddr string, data EmailSingle) (recvReport RecvReport) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	recvReport = RecvReport{
		To:        recvAddr,
		EmailData: data,
	}

	var err error
	defer func() {
		recvReport.Error = err
	}()

	// ** init the smtp client before really send
	if m.smtp == nil {
		m.smtp, err = initClient(ctx, m.Config.EmailCredential)
	}

	if err != nil {
		err = fmt.Errorf("failed to init smtp client: %w", err)
		return
	}

	if m.smtp == nil {
		err = fmt.Errorf("init smtp client still got nil client")
		return
	}

	// NOOP command to check if connection still ok
	err = m.smtp.Noop()
	if err != nil {
		err = fmt.Errorf("smtp connection is not ok: %w", err)
		return
	}

	// RSET command is for aborting already started mail transaction (tools.ietf.org/html/rfc5321#section-4.1.1.5).
	err = m.smtp.Reset()
	if err != nil {
		err = fmt.Errorf("RSET cmd failed: %w", err)
		return
	}

	// New transaction is initiated using the MAIL command (tools.ietf.org/html/rfc5321#section-4.1.1.2).
	err = m.smtp.Mail(data.SenderAddr, nil)
	if err != nil {
		err = fmt.Errorf("MAIL cmd failed: %w", err)
		return
	}

	_err := m.smtp.Rcpt(recvAddr)
	if _err != nil {
		err = fmt.Errorf("error recipient %s: %w", recvAddr, _err)
		return
	}

	// Send the email body.
	var wc io.WriteCloser
	wc, err = m.smtp.Data()
	if err != nil {
		err = fmt.Errorf("error data writer: %w", err)
		return
	}

	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	buf.WriteString(fmt.Sprintf("Recipients: %s\r\n\r\n", strings.Join(data.Recipients, ",")))
	buf.WriteString(fmt.Sprintf("%s\r\n", data.Body))

	_, err = io.Copy(wc, buf)
	if err != nil {
		err = fmt.Errorf("error data copy: %w", err)
		return
	}

	err = wc.Close()
	if err != nil {
		err = fmt.Errorf("error data close: %w", err)
		return
	}

	return
}

// Close .
// https://stackoverflow.com/questions/2468851/when-should-i-send-quit-to-smtp-server-and-how-long-should-i-keep-a-session
// https://stackoverflow.com/a/19670136/5489910
func (m *SmtpMailer) Close() error {
	if m.smtp == nil {
		return nil
	}

	var err error
	_err := m.smtp.Quit()
	if _err == nil {
		return nil
	}

	err = multierr.Append(err, fmt.Errorf("quit command error: %w", _err))
	_err = m.smtp.Close()
	if _err != nil {
		err = multierr.Append(err, fmt.Errorf("close command error: %w", _err))

		return err
	}

	return nil
}

// ----- Function here is intended to have simple function (not as method handler in a struct),
// because it will be eaiser to debug and test. In addition, we can ensure it will not use the variable that stateful.

func initClient(ctx context.Context, cred *EmailCredential) (*smtp.Client, error) {
	err := validator.New().Struct(cred)
	if err != nil {
		err = fmt.Errorf("validation on email credential error: %w", err)
		return nil, err
	}

	smtpAddr := fmt.Sprintf("%s:%d", cred.ServerHost, cred.ServerPort)

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", smtpAddr)
	if err != nil {
		err = fmt.Errorf("tcp dial error: %w", err)
		return nil, err
	}

	c, err := smtp.NewClient(conn, cred.ServerHost)
	if err != nil {
		err = fmt.Errorf("error new smtp client: %w", err)
		return nil, err
	}

	err = c.StartTLS(&tls.Config{})
	if err != nil {
		err = fmt.Errorf("error start tls: %w", err)
		return nil, err
	}

	// Set the sender and recipient first
	err = c.Auth(sasl.NewPlainClient("", cred.Username, cred.Password))
	if err != nil {
		err = fmt.Errorf("error auth: %w", err)
		return nil, err
	}

	err = c.Noop()
	if err != nil {
		err = fmt.Errorf("check smtp is not ok: %w", err)
		return nil, err
	}

	return c, nil
}
