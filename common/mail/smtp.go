package mail

import (
	"net/smtp"
	"os"
	"strconv"
)

func SendCode(recipient string, code int) error {
	login := os.Getenv("MAIL_LOGIN")
	password := os.Getenv("MAIL_PASSWORD")
	host := os.Getenv("MAIL_HOST")
	auth := smtp.PlainAuth("", login, password, host)

	from := os.Getenv("MAIL_LOGIN")
	to := []string{recipient}
	msg := []byte("To: " + recipient + "\r\n" +
		"Subject: Registration confirmation code\r\n" +
		"\r\n" +
		"Your confirmation code is " + strconv.Itoa(code) + "\r\n")

	if err := smtp.SendMail(host+":587", auth, from, to, msg); err != nil {
		return err
	}

	return nil
}

func SendPasswordResetLink(recipient string, link string) error {
	login := os.Getenv("MAIL_LOGIN")
	password := os.Getenv("MAIL_PASSWORD")
	host := os.Getenv("MAIL_HOST")
	auth := smtp.PlainAuth("", login, password, host)

	from := os.Getenv("MAIL_LOGIN")
	to := []string{recipient}
	subject := "Subject: Update password\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := `
	<html>
		<body>
			<h1>click the button below to update your password</h1>
			<a href="` + link + `">
				<button>Update password</button>
			</a>
		</body>
	</html>`

	msg := []byte(subject + mime + body)
	if err := smtp.SendMail(host+":587", auth, from, to, msg); err != nil {
		return err
	}

	return nil
}
