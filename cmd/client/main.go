package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func main() {
	smtpServer := flag.String("server", "localhost", "SMTP server hostname")
	smtpPort := flag.Int("port", 2525, "SMTP server port")
	from := flag.String("from", "sender@example.com", "Sender email address")
	to := flag.String("to", "recipient@example.com", "Recipient email address (comma-separated for multiple recipients)")
	subject := flag.String("subject", "Test Email", "Email subject")
	body := flag.String("body", "This is a test email from the Go SMTP client\n.", "Email body content")
	username := flag.String("username", "johndoe", "SMTP server username")
	password := flag.String("password", "password", "SMTP server password")
	
	flag.Parse()
	
	recipients := strings.Split(*to, ",")
	
	formattedSubject := fmt.Sprintf("Subject: %s\n", *subject)
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	message := []byte(formattedSubject + mime + *body)

	addr := fmt.Sprintf("%s:%d", *smtpServer, *smtpPort)

	// TODO: get from request
	auth := smtp.CRAMMD5Auth(*username,*password)
	err := smtp.SendMail(addr, auth, *from, recipients, message)
	if err != nil {
		log.Fatalf("Error sending email: %v", err)
	}
	
	fmt.Println("Email sent successfully!")
	fmt.Printf("From: %s\n", *from)
	fmt.Printf("To: %s\n", strings.Join(recipients, ", "))
	fmt.Printf("Subject: %s\n", *subject)
}
