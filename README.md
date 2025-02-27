# cloudflaretinyurl

## Features

- Create a short URL from a given long URL.
- Redirect a short URL to its original long URL.
- Delete a short URL.

## Requirements

- Go (>=1.16)
- Gorilla Mux package

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/yourrepo/tinyurl-api.git
   cd tinyurl-api
   ```

2. Install dependencies:

   ```sh
   go get -u github.com/gorilla/mux
   ```

3. Run the server:

   ```sh
   go run main.go
   ```

The server will start on `http://localhost:8080`.

## API Endpoints

### 1. Create Short URL

**Endpoint:** `POST /api/v1/create`

**Request Body:**

```json
{
  "long_url": "https://example.com"
}
```

**Response:**

```json
{
  "short_url": "http://localhost:8080/api/v1/abc123",
  "long_url": "https://example.com"
}
```

### 2. Redirect to Long URL

**Endpoint:** `GET /api/v1/{shortURL}`

**Description:** When a user accesses `http://localhost:8080/api/v1/abc123`, they will be redirected to `https://example.com` with a `302 Found` status.

### 3. Delete Short URL

**Endpoint:** `DELETE /api/v1/delete/{shortURL}`

**Response:** `204 No Content`


## License

MIT License


