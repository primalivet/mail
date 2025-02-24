package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	host string
	port int
}

func New(host string, port int) *Server {
	return &Server{
		host: host,
		port: port,
	}
}

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

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Minute))
	s.reply(conn, "220 Simple Go SMTP Server")
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from connection: %v", err)
			}
			return
		}

		command := strings.TrimSpace(string(buffer[:n]))

		switch {
		case strings.HasPrefix(command, "HELO") || strings.HasPrefix(command, "EHLO"):
			s.reply(conn, "250 Hello")

		case strings.HasPrefix(command, "MAIL FROM:"):
			s.reply(conn, "250 Sender OK")

		case strings.HasPrefix(command, "RCPT TO:"):
			s.reply(conn, "250 Recipient OK")

		case strings.HasPrefix(command, "DATA"):
			s.reply(conn, "354 Start mail input; end with <CRLF>.<CRLF>")
			s.handleData(conn)

		case strings.HasPrefix(command, "QUIT"):
			s.reply(conn, "221 Bye")
			return

		default:
			s.reply(conn, "500 Unknown command")
		}
	}
}

func (s *Server) handleData(conn net.Conn) {
	buffer := make([]byte, 1024)
	var message strings.Builder

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading data: %v", err)
			return
		}

		data := string(buffer[:n])
		message.WriteString(data)

		if strings.HasSuffix(message.String(), "\r\n.\r\n") {
			log.Printf("Received message:\n%s", message.String())
			s.reply(conn, "250 Message accepted for delivery")
			return
		}
	}
}

func (s *Server) reply(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message + "\r\n"))
	if err != nil {
		log.Printf("Error sending reply: %v", err)
	}
}
