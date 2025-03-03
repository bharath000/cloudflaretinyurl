# Cloudflare TinyURL - URL Shortener

## Overview
Cloudflare TinyURL is a URL shortening service that provides short, unique URLs with tracking capabilities. The system is designed with scalability, fault tolerance, and event-driven architecture.

---

## Features
- Create short URLs with a unique identifier
- Redirect short URLs to the original long URL
- Track URL clicks (last 1 minute, 24 hours, last week, all-time)
- Data persistence using PostgreSQL & Redis
- Event-driven architecture for handling click events
- Caching for fast URL resolution
- Redis queue for processing expired click events
- End-to-end testing suite for validation

---

## Prerequisites
Ensure you have the following installed:

- [Docker](https://www.docker.com/get-started)
- [Docker Compose](https://docs.docker.com/compose/install/)

---

## Setup & Installation
### 1. Clone the Repository
```sh
 git clone https://github.com/bharath000/cloudflaretinyurl
 cd cloudflaretinyurl
```

### 2. Build & Start the Services using Docker Compose
```sh
 docker-compose up --build
```
This will:
- Build the application container
- Start PostgreSQL & Redis
- Run database migrations (init-db.sql)

### 3. Verify the Containers
```sh
 docker ps
```
Ensure all services are running:
- `cloudflaretinyurl_service`
- `cloudflaretinyurl_postgres`
- `cloudflaretinyurl_redis`

---

## API Endpoints
### **Create a Short URL**
```sh
curl -X POST http://localhost:8080/api/v1/create \
     -H "Content-Type: application/json" \
     -d '{"long_url": "https://example.com"}'
```
```
Eg:
{"short_url":"http://localhost:8080/api/v1/2bK","long_url":"https://example.com","created_at":"0001-01-01T00:00:00Z"}
```

### **Redirect to Original URL**
```sh
curl -i -X GET http://localhost:8080/api/v1/{shortURL}
```
```
Eg:
HTTP/1.1 302 Found
Content-Type: text/html; charset=utf-8
Location: https://example.com
Date: Mon, 03 Mar 2025 03:19:20 GMT
Content-Length: 42

<a href="https://example.com">Found</a>.
```

### **Get Click Counts**
```sh
curl -X GET http://localhost:8080/api/v1/clicks/{shortURL}
```

```
Eg:
{"all_time":6,"last_1min":1,"last_24_hours":6,"last_week":6}
```

### **Delete a Short URL**
```sh
curl -X DELETE http://localhost:8080/api/v1/{shortURL}
```

### **Get Click Counts from Database (Fallback if Redis is Down)**
```sh
curl -X GET http://localhost:8080/api/v1/clicks_fallback/{shortURL}
```

---

## Running Tests
The system has an end-to-end (E2E) test suite.

### **Run Tests Manually (Inside Container)**
```sh
docker-compose run test_runner go test  ./test/end_to_end_test.go
docker-compose run test_runner go test -v ./test/end_to_end_test.go -run TestCreateRedirectDeleteURLsE2E
```

---

## Stopping and Cleaning Up
To stop all running containers:
```sh
docker-compose down
```

To remove all volumes and start fresh:
```sh
docker-compose down -v
```

---

## Future Improvements
- Enhance logging and monitoring
- Implement rate limiting and security features
- Scale Redis and PostgreSQL for high availability
- Advanced caching optimizations

---


