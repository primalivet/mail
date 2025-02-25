package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
	"github.com/primalivet/mail/shared/challenge"
)

// Connection state
type connectionState int

const (
	stateNew connectionState = iota
	stateGreeted
	stateAuthenticated
	stateInMailTransaction
	stateInData
)

// Connection represents a client connection to the SMTP server
type Connection struct {
	conn           net.Conn
	state          connectionState
	buffer         []byte
	from           string
	recipients     []string
	currentMessage strings.Builder
	authenticated  bool
}

// Server represents the SMTP server
type Server struct {
	host     string
	port     int
	accounts map[string]string // username -> password
}

// New creates a new SMTP Server
func New(host string, port int) *Server {
	return &Server{
		host: host,
		port: port,
		accounts: map[string]string{
			"johndoe": "password", // Add your test accounts here
		},
	}
}

// Start begins listening for SMTP connections
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()
	log.Printf("SMTP Server listening on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

// handleConnection processes a client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	c := &Connection{
		conn:   conn,
		state:  stateNew,
		buffer: make([]byte, 1024),
	}

	// Send greeting
	s.reply(conn, "220 Simple Go SMTP Server")
	c.state = stateGreeted

	for {
		// Read command
		n, err := conn.Read(c.buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		command := strings.TrimSpace(string(c.buffer[:n]))
		log.Printf("Received command: %s", command)

		// Process command based on current state
		keepGoing := s.processCommand(c, command)
		if !keepGoing {
			return
		}
	}
}

// processCommand handles a single SMTP command based on the connection state
func (s *Server) processCommand(c *Connection, command string) bool {
	if c.state == stateInData {
		return s.handleDataContent(c, command)
	}

	// Regular command processing
	lowerCmd := strings.ToLower(command)
	switch {
	case strings.HasPrefix(lowerCmd, "helo") || strings.HasPrefix(lowerCmd, "ehlo"):
		s.handleHelo(c)
		return true

	case strings.HasPrefix(lowerCmd, "auth cram-md5"):
		s.handleCramMD5Auth(c)
		return true

	case strings.HasPrefix(lowerCmd, "mail from:"):
		if c.authenticated {
			from := extractEmail(command, "mail from:")
			c.from = from
			s.reply(c.conn, "250 Sender OK")
			c.state = stateInMailTransaction
		} else {
			s.reply(c.conn, "530 Authentication required")
		}
		return true

	case strings.HasPrefix(lowerCmd, "rcpt to:"):
		if c.state >= stateInMailTransaction {
			recipient := extractEmail(command, "rcpt to:")
			c.recipients = append(c.recipients, recipient)
			s.reply(c.conn, "250 Recipient OK")
		} else {
			s.reply(c.conn, "503 Bad sequence of commands")
		}
		return true

	case strings.HasPrefix(lowerCmd, "data"):
		if c.state >= stateInMailTransaction && len(c.recipients) > 0 {
			s.reply(c.conn, "354 Start mail input; end with <CRLF>.<CRLF>")
			c.state = stateInData
			c.currentMessage.Reset()
		} else {
			s.reply(c.conn, "503 Bad sequence of commands")
		}
		return true

	case strings.HasPrefix(lowerCmd, "quit"):
		s.reply(c.conn, "221 Bye")
		return false

	default:
		log.Printf("Unknown command: %s", command)
		s.reply(c.conn, "500 Unknown command")
		return true
	}
}

// handleHelo processes HELO/EHLO commands
func (s *Server) handleHelo(c *Connection) {
	s.reply(c.conn, "250-localhost")
	s.reply(c.conn, "250-SIZE 10485760")
	s.reply(c.conn, "250-AUTH CRAM-MD5")
	s.reply(c.conn, "250 8BITMIME")
	c.state = stateGreeted
}

// handleCramMD5Auth processes CRAM-MD5 authentication
func (s *Server) handleCramMD5Auth(c *Connection) {
	// Generate challenge
	originalChallenge, encodedChallenge := challenge.Generate(s.host)
	log.Printf("Generated challenge: %s", originalChallenge)

	// Send challenge
	s.reply(c.conn, fmt.Sprintf("334 %s", encodedChallenge))

	// Read response
	authN, err := c.conn.Read(c.buffer)
	if err != nil {
		log.Printf("Error reading auth response: %v", err)
		return
	}

	// Process response
	clientResponse := strings.TrimSpace(string(c.buffer[:authN]))
	log.Printf("Received auth response: %s", clientResponse)

	// Verify credentials
	if challenge.Verify(s.accounts, originalChallenge, clientResponse) {
		s.reply(c.conn, "235 Authentication successful")
		c.authenticated = true
	} else {
		s.reply(c.conn, "535 Authentication failed")
	}
}

// handleDataContent processes the content of an email
func (s *Server) handleDataContent(c *Connection, data string) bool {
	// Append new data to buffer
	c.currentMessage.WriteString(data)
	buffer := c.currentMessage.String()

	if strings.Contains(buffer, "\r\n.\r\n") ||
		strings.Contains(buffer, "\n.\r\n") ||
		strings.Contains(buffer, "\r\n.\n") ||
		strings.Contains(buffer, "\n.\n") ||
		strings.HasSuffix(buffer, "\r\n.") ||
		strings.HasSuffix(buffer, "\n.") {

		// Look for the standalone dot on its own line
		// End of data found
		message := strings.Split(buffer, "\r\n.\r\n")[0]

		// Process message (with dot unstuffing)
		processedMessage := s.unstuffDots(message)
		log.Printf("Received message:\n%s", processedMessage)

		s.reply(c.conn, "250 Message accepted for delivery")
		c.state = stateGreeted
		return true
	}

	return true
}

// Unstuff dots (remove leading dots that were added according to SMTP rules)
func (s *Server) unstuffDots(message string) string {
	lines := strings.Split(message, "\r\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "..") {
			lines[i] = line[1:] // Remove one dot
		}
	}
	return strings.Join(lines, "\r\n")
}

// Helper functions

// reply sends a response to the client
func (s *Server) reply(conn net.Conn, message string) {
	log.Printf("Sending: %s", message)
	_, err := conn.Write([]byte(message + "\r\n"))
	if err != nil {
		log.Printf("Error sending reply: %v", err)
	}
}

// extractEmail extracts the email address from a command
func extractEmail(command, prefix string) string {
	address := strings.TrimPrefix(command, prefix)
	address = strings.Trim(address, "<>")
	address = strings.TrimSpace(address)
	return address
}
