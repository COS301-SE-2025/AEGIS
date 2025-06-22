package verifyemail

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(to string, token string) error {
	verifyURL := fmt.Sprintf("https://capstone-aegis.dns.net.za/verify-email?token=%s", token)
	subject := "Welcome to AEGIS — Verify Your Email"
	body := fmt.Sprintf(`Hello,

You have recently been registered on the AEGIS platform. Please verify your email by clicking the link below:

%s

If you did not expect this email, you can ignore it.

- The AEGIS Team`, verifyURL)

	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s",
		os.Getenv("SMTP_FROM"), to, subject, body)

	auth := smtp.PlainAuth("", os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), os.Getenv("SMTP_HOST"))
	addr := fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))

	return smtp.SendMail(addr, auth, os.Getenv("SMTP_FROM"), []string{to}, []byte(msg))
}
