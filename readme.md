# Go Proxy Server

This is a simple HTTP/HTTPS proxy server implemented in Go. It allows you to intercept and manipulate HTTP requests and responses.

## Features

- **HTTP and HTTPS Proxy**: Handles both HTTP and HTTPS requests.
- **Request Logging**: Logs incoming requests and responses to the console.
- **Request Filtering**: Optionally blocks requests to specified URLs.
- **Caching**: Caches responses for faster retrieval.

## Requirements

- Go (version 1.16 or higher)

## Installation

1. Clone the repository:

```bash
   git clone https://github.com/itsfuad/pokeproxy.git
```bash
   cd pokeproxy
```

2. Run the server:

```bash
   go run main.go
```

3. Configure your browser to use the proxy server. Set the proxy server to `localhost:8080`.
4. To block requests to specific URLs, add the URLs to the `blockedURLs` slice in the `main.go` file.