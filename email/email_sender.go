package email

import (
	"fmt"

	"gopkg.in/mail.v2"
)

func SendResetEmail(to, token string) error {
	resetURL := fmt.Sprintf("http://localhost:5173/password-reset?token=%s&email=%s", token, to)

	m := mail.NewMessage()
	m.SetHeader("From", "noreply@gm.gg")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Reset Your Password")
	m.SetBody("text/plain", "Click to reset: "+resetURL)

	d := mail.NewDialer("localhost", 1025, "", "")

	return d.DialAndSend(m)
}
