package mailclient

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	t.Run("dry run", func(t *testing.T) {

		ctx := context.TODO()
		cfg := &SmtpMailerConfig{
			EmailCredential: &EmailCredential{
				Protocol:     "smtp",
				ServerHost:   "smtp.gmail.com",
				ServerPort:   465,
				AuthIdentity: "",
				Username:     "xxx@gmail.com",
				Password:     "---",
			},
		}

		client, err := NewSmtp(cfg)
		assert.NotNil(t, client)
		assert.NoError(t, err)

		emailData := EmailTemplate{
			TrackingID: "uid",
			Recipients: []string{"alice@example.com", "bob@example.com"},
			// see https://stackoverflow.com/a/26152110/5489910
			Subject: `Hi {{ with (index .recipient_map .Vars.Recipient) }}{{ .name }}{{ end }} use your promo code {{ with (index .recipient_map .Vars.Recipient) }}{{ .promo_code }}{{ end }} today!`,
			SubjectData: map[string]interface{}{
				"recipient_map": map[string]interface{}{
					"alice@example.com": map[string]string{
						"name":       "Alice",
						"promo_code": "ALC1",
					},
					"bob@example.com": map[string]string{
						"name":       "Bob",
						"promo_code": "BOB1",
					},
				},
			},
			Body:        "Dear our valued customer,\nBy this email we want to thank you for your registration at BlaBlaBla service!",
			BodyData:    map[string]interface{}{},
			Attachments: nil,
		}

		report := client.Send(ctx, emailData, DryRun())
		assert.NotEmpty(t, report)

		for _, recvReport := range report.RecvReports {
			t.Logf("%+v\n", recvReport.EmailData)
		}

	})

}
