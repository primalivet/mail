# mail

## Start server
```
go run cmd/server/main.go
```

## Telnet example payload
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

## Client example payload

```
# Using all flags
go run cmd/client/main.go -server=localhost -port=2525 -from="john@doe.com" -to="jane@doe.com" -subject="Hello, World" - body="This is a custom test email"

# Using default flags
go run cmd/client/main.go
```
