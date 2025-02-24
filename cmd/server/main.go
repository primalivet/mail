package main

import (
    "fmt"
    "io"
    "log"
    "net"
    "strings"
    "time"
)

type EmailServer struct {
    host string
    port int
}

// Create a new email server instance
func NewEmailServer(host string, port int) *EmailServer {
    return &EmailServer{
        host: host,
        port: port,
    }
}

// Start the SMTP server
func (s *EmailServer) Start() error {
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

// Handle individual SMTP connections
func (s *EmailServer) handleConnection(conn net.Conn) {
    defer conn.Close()

    // Set connection timeout
    conn.SetDeadline(time.Now().Add(5 * time.Minute))

    // Send greeting
    s.reply(conn, "220 Simple Go SMTP Server")

    // Create a buffer for reading commands
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
        
        // Handle SMTP commands
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

// Handle email data
func (s *EmailServer) handleData(conn net.Conn) {
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

        // Check for end of message
        if strings.HasSuffix(message.String(), "\r\n.\r\n") {
            log.Printf("Received message:\n%s", message.String())
            s.reply(conn, "250 Message accepted for delivery")
            return
        }
    }
}

// Send SMTP reply
func (s *EmailServer) reply(conn net.Conn, message string) {
    _, err := conn.Write([]byte(message + "\r\n"))
    if err != nil {
        log.Printf("Error sending reply: %v", err)
    }
}

func main() {
    server := NewEmailServer("localhost", 2525)
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
