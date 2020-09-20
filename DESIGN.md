### Error model
Actionable failure conditions are reported in `4xx` responses as JSON in
the response body in the following format:

```json
"errors": [
	"code": <signed integer>,
	"message": "<string>",
	"detail": {...}
]
```

Clients should retry requests with exponential back-off for all
transient errors.
