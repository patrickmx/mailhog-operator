package main

import (
	"log"
	"net/smtp"
)

// GOCACHE="/tmp/" go run testmail.go
func main() {
	// Configuration
	from := "testmail@localhost.local"
	password := "anything"
	to := []string{"mailhog@localhost.local"}
	smtpHost := "localhost"
	smtpPort := "1025"

	message := []byte("My super secret message.")

	// Create authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send actual message
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("success")
	}
}
