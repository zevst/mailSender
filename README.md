# MailSender

```go
package main

import (
	"html/template"

	"github.com/mylockerteam/mailSender"
	"gopkg.in/gomail.v2"
)

const ess = "<your ess in format host:port;username;password>"

func GetMailSender() mailSender.AsyncSender {
	return mailSender.Create(&mailSender.Sender{
		Channel: make(chan mailSender.Message, 1),
		Closer:  mailSender.GetCloser(gomail.NewDialer(mailSender.ParseEss(ess))),
	})
}

func main() {
	sender := GetMailSender()
	m := gomail.NewMessage()
	tpl, _ := template.New("test").Parse("<div>Hi, {{ .Name }}!</div>")
	m.SetHeader("From", "Example <no-reply@example.com>")
	m.SetHeader("Bcc", "<test@example.com>")
	m.SetHeader("Subject", "Hello!")
	sender.SendAsync(mailSender.Message{
		Message:  m,
		Template: tpl,
		Data:     mailSender.EmailData{"Name": "Tester"},
	})
	// ...
}
```