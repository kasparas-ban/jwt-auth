# JWT-based auth server

To start the server add .env file in the format as shown below and run

`go run main.go`

## .env file example

```
HOST_SERVER=smtp.example.com
EMAIL_PORT=465
EMAIL_DOMAIN=example.com
EMAIL_USER=example@example.com
EMAIL_PASS=example

JWT_KEY=example

PORT=3001
```

---

## TODO list

- [ ] Add appropriate headers against XSS attacks
- [ ] Add tests
- [ ] Add input validation for login
- [ ] Add input validation for registration
- [ ] Add request-rate limiting against DDOS attacks
- [ ] Add Authorization header
