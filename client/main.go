package main

import (
	"flag"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func main() {
	// Define command line flags
	smtpServer := flag.String("server", "localhost", "SMTP server hostname")
	smtpPort := flag.Int("port", 2525, "SMTP server port")
	from := flag.String("from", "sender@example.com", "Sender email address")
	to := flag.String("to", "recipient@example.com", "Recipient email address (comma-separated for multiple recipients)")
	subject := flag.String("subject", "Test Email", "Email subject")
	body := flag.String("body", "This is a test email from the Go SMTP client.", "Email body content")
	
	// Parse the command line arguments
	flag.Parse()
	
	// Split recipients if multiple are provided
	recipients := strings.Split(*to, ",")
	
	// Format the email content
	formattedSubject := fmt.Sprintf("Subject: %s\n", *subject)
	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	message := []byte(formattedSubject + mime + *body)

	// Connect to the server
	addr := fmt.Sprintf("%s:%d", *smtpServer, *smtpPort)
	
	// Since our server doesn't have authentication yet, we pass empty auth
	err := smtp.SendMail(addr, nil, *from, recipients, message)
	if err != nil {
		log.Fatalf("Error sending email: %v", err)
	}
	
	fmt.Println("Email sent successfully!")
	fmt.Printf("From: %s\n", *from)
	fmt.Printf("To: %s\n", strings.Join(recipients, ", "))
	fmt.Printf("Subject: %s\n", *subject)
}
