package verifyemail

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
)

func SendVerificationEmail(to string, token string) error {
	fmt.Println("📧 Sending verification email to:", to)
	verifyURL := fmt.Sprintf("http://localhost:5173/verify-email?token=%s", token)
	subject := "Welcome to AEGIS — Verify Your Email"
	body := fmt.Sprintf(`Hello,

You have recently been registered on the AEGIS platform. Please verify your email by clicking the link below:

%s

If you did not expect this email, you can ignore it.

- The AEGIS Team`, verifyURL)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s",
		os.Getenv("SMTP_FROM"), to, subject, body)

	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	addr := net.JoinHostPort(host, port)

	auth := smtp.PlainAuth("", os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"), host)

	// TLS config with server name for validation
	tlsconfig := &tls.Config{
		ServerName: host,
	}

	// Connect to SMTP server
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth error: %w", err)
	}

	// Set the sender and recipient first
	if err = client.Mail(os.Getenv("SMTP_FROM")); err != nil {
		return fmt.Errorf("failed to set mail sender: %w", err)
	}

	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set mail recipient: %w", err)
	}

	// Data
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send data command: %w", err)
	}

	_, err = wc.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	err = wc.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}
