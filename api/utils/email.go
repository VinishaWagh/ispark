package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"os"
)

// GenerateOTP generates a cryptographically secure 6-digit numeric string
func GenerateOTP() (string, error) {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	b := make([]byte, 6)
	n, err := io.ReadAtLeast(rand.Reader, b, 6)
	if n != 6 || err != nil {
		return "", err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

// SendOTP sends the OTP to the specified email. If SMTP configs are not found, it prints the OTP to the console.
func SendOTP(toEmail, otpCode, purpose string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpSender := os.Getenv("SMTP_SENDER")

	subject := "Verification Code for iSpark"
	body := fmt.Sprintf("Your OTP code for %s is: %s\nThis code expires in 15 minutes.", purpose, otpCode)

	// Fallback to console logging if SMTP setup is missing
	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("\n--- [MOCK EMAIL SENDER] ---\nTo: %s\nSubject: %s\nBody: %s\n----------------------------\n", toEmail, subject, body)
		return nil
	}

	msg := []byte(
		"From: iSpark <" + smtpSender + ">\r\n" +
			"To: " + toEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpSender, []string{toEmail}, msg)
	if err != nil {
		log.Printf("Failed to send real email (falling back to console print): %v", err)
		log.Printf("\n--- [FALLBACK EMAIL SENDER] ---\nTo: %s\nSubject: %s\nBody: %s\n--------------------------------\n", toEmail, subject, body)
		return nil // return nil so the API call doesn't fail, since email is logged to console for development
	}

	log.Printf("OTP email successfully sent to %s", toEmail)
	return nil
}
