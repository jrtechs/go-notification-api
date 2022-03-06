package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"

	"github.com/jrtechs/go-notification-api/conf"

	gomail "gopkg.in/mail.v2"
)

var smptServer string
var fromEmail string
var emailPassword string

func InitConfig() error {

	var envExists bool

	fromEmail, envExists = os.LookupEnv("email")
	if !envExists {
		conf.Logger.Println("Unable to fetch 'email' environment variable")
		return errors.New("")
	}
	conf.Logger.Println("Setting from email to: ", fromEmail)

	smptServer, envExists = os.LookupEnv("smtp")
	if !envExists {
		conf.Logger.Println("Unable to fetch 'smtp' environment variable")
		return errors.New("")
	}
	conf.Logger.Println("Setting smtp server to: ", smptServer)

	emailPassword, envExists = os.LookupEnv("password")
	if !envExists {
		conf.Logger.Println("Unable to fetch 'password' environment variable")
		return errors.New("")
	}
	conf.Logger.Println("Loaded email password")
	_ = emailPassword

	return nil
}

func SendEmail(destinationEmail string, subject string, message string) error {

	conf.Logger.Printf("Sending email from %s and to %s with message %s", fromEmail, destinationEmail, message)

	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", fromEmail)

	// Set E-Mail receivers
	m.SetHeader("To", destinationEmail)

	// Set E-Mail subject
	m.SetHeader("Subject", subject)

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/plain", message)

	// Settings for SMTP server
	d := gomail.NewDialer(smptServer, 587, fromEmail, emailPassword)

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
