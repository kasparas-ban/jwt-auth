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
- [ ] Implement max-retries for login

## Things to be aware of

- Validate `content-type` on request `Accept` header (Content Negotiation) to allow only your supported format (e.g., `application/xml`, `application/json`, etc.) and respond with `406 Not Acceptable response` if not matched.
- Use response type: `application/json` and `charset=utf-8`
- Use appropriate headers
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: deny`
  - `X-XSS-Protection: 0`
  - `Cache-Control: 'no-store'`
  - `Content-Security-Policy: default-src 'none' frame-ancestors 'none'; sandbox`
  - `Server: ''`

## Request handling in order

- (Rate Limiter) If exceeds server computational capabilities, ignore the request
- Deal with CORS if needed
  - Check if it's preflight, set Origin related headers
- Enforce that POST requests must be of type `application/json` (otherwise return 415)
- Set response headers
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: deny`
  - `X-XSS-Protection: 0`
  - `Cache-Control: 'no-store'`
  - `Content-Security-Policy: default-src 'none' frame-ancestors 'none'; sandbox`
  - `Server: ''`
- Authenticate incoming response
- Validate token
- Enable logging (log incomming request as well as the outgoing response before it's sent)
- API handlers go here
- Respond with `Not Found 404`
