// Package mail provides e-mail sending faculties.
package mail

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/aichaos/silhouette/webapp/config"
	"github.com/aichaos/silhouette/webapp/log"
	"github.com/microcosm-cc/bluemonday"
	"gopkg.in/gomail.v2"
)

// Message configuration.
type Message struct {
	To       string
	ReplyTo  string
	Subject  string
	Template string // path relative to the templates dir, e.g. "email/verify_email.html"
	Data     map[string]interface{}
}

// Send an email.
func Send(msg Message) error {
	conf := config.Current.Mail

	// Verify configuration.
	if !conf.Enabled {
		return errors.New(
			"Email sending is not configured for this app. Please contact the website administrator about this error.",
		)
	} else if conf.Host == "" || conf.Port == 0 || conf.From == "" {
		return errors.New(
			"Email settings are misconfigured for this app. Please contact the website administrator about this error.",
		)
	}

	// Get and render the template to HTML.
	var html bytes.Buffer
	tmpl, err := template.New(msg.Template).ParseFiles(config.TemplatePath + "/" + msg.Template)
	if err != nil {
		return err
	}

	// Execute the template.
	err = tmpl.ExecuteTemplate(&html, "content", msg)
	if err != nil {
		return fmt.Errorf("Mail template execute error: %s", err)
	}

	// Condense the HTML down into the plaintext version.
	rawLines := strings.Split(
		bluemonday.StrictPolicy().Sanitize(html.String()),
		"\n",
	)
	var lines []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		lines = append(lines, line)
	}
	plaintext := strings.Join(lines, "\n\n")

	// Prepare the e-mail!
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", config.Title, conf.From))
	m.SetHeader("To", msg.To)
	if msg.ReplyTo != "" {
		m.SetHeader("Reply-To", msg.ReplyTo)
	}
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/plain", plaintext)
	m.AddAlternative("text/html", html.String())

	// Deliver asynchronously.
	log.Info("mail.Send: %s (%s) to %s", msg.Subject, msg.Template, msg.To)
	d := gomail.NewDialer(conf.Host, conf.Port, conf.Username, conf.Password)
	go func() {
		if err := d.DialAndSend(m); err != nil {
			log.Error("mail.Send: %s", err.Error())
		}
	}()

	return nil
}
