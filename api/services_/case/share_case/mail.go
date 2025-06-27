package sharecase

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendCaseShareEmail(to string, token string) error {
	viewURL := fmt.Sprintf("https://capstone-aegis.dns.net.za/view-case?token=%s", token)

	subject := "AEGIS: A Case Has Been Shared With You"
	body := fmt.Sprintf(`Hello,

You have been invited to collaborate on a case on the AEGIS platform.

Click the link below to access the case:
%s

If this was not expected, please contact the admin.

– The AEGIS Team`, viewURL)

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s",
		os.Getenv("SMTP_FROM"), to, subject, body)

	auth := smtp.PlainAuth("", os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), os.Getenv("SMTP_HOST"))
	addr := fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))

	return smtp.SendMail(addr, auth, os.Getenv("SMTP_FROM"), []string{to}, []byte(msg))
}
