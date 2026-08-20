# Internal error catalogue

The challenge's public `400` and `500` schemas contain only a human-friendly `message`. Numeric codes are therefore an internal operational contract: they make logs searchable without breaking the supplied API.

| Code | Meaning | Public status |
| ---: | --- | ---: |
| 1001 | Booking list is empty | 400 |
| 1002 | Booking count or body size exceeds its limit | 400 |
| 1003 | Media type or JSON payload is invalid | 400 |
| 1101 | Request ID is not a valid UUID v4 | 400 |
| 1102 | Request ID is duplicated in the payload | 400 |
| 1201 | Check-in date is invalid | 400 |
| 1301 | Nights are outside 1–365 | 400 |
| 1401 | Selling rate is not positive | 400 |
| 1501 | Margin is outside 0–100 | 400 |
| 9000 | Unexpected or invariant-breaking internal failure | 500 |

Clients should depend on the HTTP status and message contract, not these internal codes. Operators can correlate the numeric code with `X-Request-ID` and structured application logs.
