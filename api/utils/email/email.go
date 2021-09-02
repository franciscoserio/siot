package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
)

func SendConfirmationEmail(to []string, subject string, firstName string, lastName string, InvitationToken string) {

	// Sender data
	from := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	// smtp server configuration
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	t, _ := template.ParseFiles("api/utils/email/templates/confirmation_email_template.html")

	var body bytes.Buffer

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body.Write([]byte(fmt.Sprintf("Subject: "+subject+" \n%s\n\n", mimeHeaders)))

	confirmationUrl := os.Getenv("SERVER_URL") + "/api/users/confirmation?confirmation_token=" + InvitationToken

	t.Execute(&body, struct {
		FirstName       string
		LastName        string
		ConfirmationUrl string
	}{
		FirstName:       firstName,
		LastName:        lastName,
		ConfirmationUrl: confirmationUrl,
	})

	// Sending email.
	smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, body.Bytes())
}
