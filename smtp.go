package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"strings"

	"github.com/bradfitz/go-smtpd/smtpd" // Subdir
)

type envelope struct {
	*smtpd.BasicEnvelope

	msg []byte
}

func (e *envelope) AddRecipient(rcpt smtpd.MailAddress) error {
	if strings.HasPrefix(rcpt.Email(), "bad@") {
		return errors.New("we don't send email to bad@")
	}
	return e.BasicEnvelope.AddRecipient(rcpt)
}

func (e *envelope) WriteLine(line []byte) error {
	e.msg = append(e.msg, line...)

	return nil
}

func (e *envelope) Close() error {
	log.Println(string(e.msg))

	r := bytes.NewReader(e.msg)
	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}

	header := m.Header
	fmt.Println("Date:", header.Get("Date"))
	fmt.Println("From:", header.Get("From"))
	fmt.Println("To:", header.Get("To"))
	fmt.Println("Subject:", header.Get("Subject"))

	body, err := ioutil.ReadAll(m.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", body)

	return nil
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
	log.Printf("dotplan: new mail from %q", from)
	return &envelope{
		BasicEnvelope: new(smtpd.BasicEnvelope),
	}, nil
}

func startSMTP() {
	s := &smtpd.Server{
		Addr:      "0.0.0.0:2525",
		OnNewMail: onNewMail,
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
