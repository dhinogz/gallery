package main

import (
	"os"

	"github.com/go-mail/mail/v2"
)

const host = "sandbox.smtp.mailtrap.io"
const port = 587
const username = "a"
const password = "w"

func main() {
	from := "test@gallery.com"
	to := "davidhinojosagzz@gmail.com"
	subject := "Test email"
	plaintext := "Email body"
	html := `<h1>Hello there buddy!</h1><p>This is the email</p><p>Hope you enjoy it</p>`

	msg := mail.NewMessage()
	msg.SetHeader("To", to)
	msg.SetHeader("From", from)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", plaintext)
	msg.AddAlternative("text/html", html)
	msg.WriteTo(os.Stdout)

	dialer := mail.NewDialer(host, port, username, password)

	err := dialer.DialAndSend(msg)
	if err != nil {
		panic(err)
	}
}
