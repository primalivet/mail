# mail

## Start server
```
go run server/main.go
```

## Example telnet payload
```
HELO example.com
MAIL FROM:<sender@example.com>
RCPT TO:<recipient@example.com>
DATA
Subject: Test Email

This is a test message.
.
QUIT
```
