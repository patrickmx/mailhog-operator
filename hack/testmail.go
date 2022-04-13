package main

import (
	"log"
	"net/smtp"
)

// GOCACHE="/tmp/" go run testmail.go
func main() {
	message := []byte("hello mail")
	auth := smtp.PlainAuth("", "user", "password", "localhost")
	if err := smtp.SendMail("localhost:1025", auth, "sender@localhost.local", []string{"receiver@localhost.local"}, message); err != nil {
		log.Fatal(err)
	} else {
		log.Println("success")
	}
}
