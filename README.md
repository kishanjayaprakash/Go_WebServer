# Simple Web Server with Rate Limiting

A lightweight HTTP web server written in Go with IP-based rate limiting and full user management.

## Features

- IP-based rate limiting (15 requests)
- Track last seen timestamp per IP
- User management (Create, Read, Delete)
- Thread-safe visitor tracking with mutex

---

## Rate Limiting

Every incoming request is tracked by IP address.

| Rule | Value |
|------|-------|
| Max requests | 15 per IP  |
| Exceeded response | 429 Too Many Requests |

How it works:
- First request from an IP creates a visitor entry
- Each request increments the counter
- Counter resets after 1 minute
- Exceeding 15 requests returns 429 Too Many Requests

---

## API Endpoints

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /users | Get all users |
| POST | /users | Create a new user |
| DELETE | /users/{id} | Delete a user by ID |

### Create User

POST /users
Content-Type: application/json

{
    "id": 1,
    "name": "Kishan",
    "email": "kishan@example.com"
}

Response:
{
    "id": 1,
    "name": "Kishan",
    "email": "kishan@example.com",
    "last_seen": "2026-06-10T09:53:02Z"
}

### Get All Users

GET /users

Response:
[
    {
        "id": 1,
        "name": "Kishan",
        "email": "kishan@example.com",
        "last_seen": "2026-06-10T09:53:02Z"
    },
    {
        "id": 2,
        "name": "John",
        "email": "john@example.com",
        "last_seen": "2026-06-10T09:55:00Z"
    }
]

### Delete User

DELETE /users/1

Response: 204 No Content

---

## Visitor Tracking

Each IP is tracked with the following details:

| Field | Description |
|-------|-------------|
| IP Address | key used to identify the visitor |
| Count | number of requests made |
| LastSeen | timestamp of the last request |

After 1 minute the counter resets and the IP can make 15 new requests.

---

## Project Structure

.
├── main.go            → entry point and server setup
├── ratelimit.go       → rate limiting middleware
├── handler.go         → user handlers
├── domain.go          → user struct and storage
└── go.mod

---

## Getting Started

### Prerequisites
- Go 1.21+

### Install dependencies

go mod tidy

### Run the server

go run .

Server starts on http://localhost:8080

---

## Testing

### Create a user
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"id": 1, "name": "Kishan", "email": "kishan@example.com"}'

### Get all users
curl http://localhost:8080/users

### Delete a user
curl -X DELETE http://localhost:8080/users/1

### Test rate limiting
for i in $(seq 1 16); do curl http://localhost:8080/users; done

Expected output on 16th request:
HTTP 429 Too Many Requests
Too many requests

---

## Tech Stack

| Technology | Purpose |
|------------|---------|
| Go standard library | core language |
| net/http | HTTP server |
| sync.Mutex | thread safe visitor map |
| time | request window tracking |

---

## How Rate Limiting Works Internally

Request comes in
      ↓
Extract IP address
      ↓
IP exists in map?
  ↓ NO                    ↓ YES
create visitor         increment count
count = 1              check last seen
      ↓                    ↓
allow request        1 minute passed?
                      ↓ YES     ↓ NO
                   reset      count > 15?
                   count      ↓ YES  ↓ NO
                      ↓     block   allow
                   allow     429    request
                   request
