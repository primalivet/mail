# mail

## Start server
```
go run cmd/server/main.go
```

## Client example payload

```
# Using all flags
go run cmd/client/main.go -server=localhost -port=2525 -from="john@doe.com" -to="jane@doe.com" -subject="Hello, World" -body="This is a custom test email" -username="johndoe" -password="password"

# Using default flags
go run cmd/client/main.go
```
