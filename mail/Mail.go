package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
)

//baun.bitirme@gmail.com
//1q2w3e4*
//Google uygulama adı = Bitirme
//uygulama şifresi = = vuca awrl whcv snjl

type Mail struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func ParseTemplate(templateFileName string, data interface{}) (Mail, error) {
	var m Mail
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return Mail{}, err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return Mail{}, err
	}
	m.Body = buf.String()
	return m, nil
}

func SendEmail(to string, code string) error {
	data := struct {
		Code string `json:"code"`
	}{
		Code: code,
	}

	m, err := ParseTemplate("/app/mail/Template.html", data)
	if err != nil {
		return fmt.Errorf("Mail template parse edilemedi.%w", err)
	}

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + m.Subject + "!\n"
	msg := []byte(subject + mime + "\n" + m.Body)

	plain := smtp.PlainAuth("", os.Getenv("MAIL_FROM"), os.Getenv("MAIL_PASSWORD"), os.Getenv("MAIL_HOST"))

	if err := smtp.SendMail(os.Getenv("MAIL_HOST_ADDRESS"), plain, os.Getenv("MAIL_FROM"), []string{to}, msg); err != nil {
		return fmt.Errorf("İlgili mail gönderilemedi. Hata: %v", err)
	}
	return nil
}
