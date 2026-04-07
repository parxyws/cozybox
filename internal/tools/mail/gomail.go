package mail

import (
	"github.com/parxyws/cozybox/internal/config"
	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer *gomail.Dialer
}

func NewGoMailDialer(cfg *config.Config) *gomail.Dialer {
	return gomail.NewDialer(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password)
}

func NewMailer(dialer *gomail.Dialer) *Mailer {
	return &Mailer{dialer: dialer}
}

func (m *Mailer) SendOTP(to string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", "[EMAIL_ADDRESS]")
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return m.dialer.DialAndSend(message)
}

func (m *Mailer) SendWelcomeEmail(to string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", "[EMAIL_ADDRESS]")
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return m.dialer.DialAndSend(message)
}

func (m *Mailer) SendResetPassword(to string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", "[EMAIL_ADDRESS]")
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return m.dialer.DialAndSend(message)
}

func (m *Mailer) SendChangeEmailRequest(to string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", "[EMAIL_ADDRESS]")
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return m.dialer.DialAndSend(message)
}

func (m *Mailer) SendChangeEmailConfirmation(to string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", "[EMAIL_ADDRESS]")
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)
	return m.dialer.DialAndSend(message)
}
