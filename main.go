package main

import (
    "io"
    "log"
    "net"
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "sync"
    "time"
    "os"
    "bufio"
)

// BlockedURLs is a list of URLs that the proxy will block.
var BlockedURLs = []string{} // read from blockedURLs

// CachedResponse stores the HTTP response and its expiration time.
type CachedResponse struct {
    Response   *http.Response
    Expiration time.Time
}

// Cache to store responses.
var cache = make(map[string]*CachedResponse)
var cacheMutex = sync.Mutex{}

// Cache duration.
const cacheDuration = 10 * time.Minute

// ProxyHandler handles incoming requests and forwards them to the target server.
func ProxyHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the URL from the request.
    targetURL, err := url.Parse(r.RequestURI)
    if err != nil {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }

    // Check if the URL is in the blocked list.
    for _, blocked := range BlockedURLs {
        if strings.Contains(targetURL.Host, blocked) {
            http.Error(w, "Access to this URL is blocked", http.StatusForbidden)
            log.Printf("Blocked request to: %s", targetURL)
            return
        }
    }

    // Log the incoming request.
    log.Printf("Received request from %s for %s", r.RemoteAddr, targetURL)

    // Check the cache.
    cacheMutex.Lock()
    cachedResponse, found := cache[targetURL.String()]
    cacheMutex.Unlock()

    if found && cachedResponse.Expiration.After(time.Now()) {
        log.Printf("Serving cached response for: %s", targetURL)
        copyResponse(w, cachedResponse.Response)
        return
    }

    // Create a reverse proxy.
    proxy := httputil.NewSingleHostReverseProxy(targetURL)

    // Modify the response before sending it to the client.
    proxy.ModifyResponse = func(resp *http.Response) error {
        log.Printf("Response status: %s", resp.Status)

        // Cache the response.
        cacheMutex.Lock()
        cache[targetURL.String()] = &CachedResponse{
            Response:   cloneResponse(resp),
            Expiration: time.Now().Add(cacheDuration),
        }
        cacheMutex.Unlock()
        return nil
    }

    // Serve the request using the proxy.
    proxy.ServeHTTP(w, r)
}

// copyResponse copies the cached response to the ResponseWriter.
func copyResponse(w http.ResponseWriter, resp *http.Response) {
    for k, v := range resp.Header {
        for _, vv := range v {
            w.Header().Add(k, vv)
        }
    }
    w.WriteHeader(resp.StatusCode)
    io.Copy(w, resp.Body)
}

// cloneResponse creates a deep copy of the HTTP response.
func cloneResponse(resp *http.Response) *http.Response {
    body, _ := io.ReadAll(resp.Body)
    resp.Body = io.NopCloser(io.MultiReader(strings.NewReader(string(body))))
    return &http.Response{
        Status:           resp.Status, // Clone the response status.
        StatusCode:       resp.StatusCode,
        Proto:            resp.Proto,
        ProtoMajor:       resp.ProtoMajor, // Clone the protocol version.
        ProtoMinor:       resp.ProtoMinor, // Clone the protocol version.
        Header:           resp.Header,
        Body:             io.NopCloser(strings.NewReader(string(body))), // Clone the body.
        ContentLength:    resp.ContentLength,
        TransferEncoding: resp.TransferEncoding, // Clone the transfer encoding.
        Close:            resp.Close,
        Uncompressed:     resp.Uncompressed,
        Trailer:          resp.Trailer,
        Request:          resp.Request,
        TLS:              resp.TLS, // Clone the TLS connection state.
    }
}

// handleHTTPS handles HTTPS connections.
func handleHTTPS(w http.ResponseWriter, r *http.Request) {
    destConn, err := net.Dial("tcp", r.Host)
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    hijacker, ok := w.(http.Hijacker) // Hijack the connection.
    if !ok {
        http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
        return
    }
    clientConn, _, err := hijacker.Hijack() 
    if err != nil {
        http.Error(w, err.Error(), http.StatusServiceUnavailable)
        return
    }
    clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n")) // Send a 200 OK response.
    go transfer(destConn, clientConn)
    go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
    defer destination.Close()
    defer source.Close()
    io.Copy(destination, source)
}

// readBlockedURLs reads the list of blocked URLs from a file.
func readBlockedURLs() []string {
    blockedURLs := []string{}
    file, err := os.Open("blockedURLs")
    if err != nil {
        log.Println("Blocked URLs file not found")
        return blockedURLs
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        blockedURLs = append(blockedURLs, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        log.Println("Error reading blocked URLs file")
    }
    return blockedURLs
}

func main() {

    //read file for blocked urls
    BlockedURLs = readBlockedURLs()

    // Handle HTTP requests
    http.HandleFunc("/", ProxyHandler)
    // Handle HTTPS requests
    http.HandleFunc("/https", handleHTTPS)

    log.Println("Proxy server running on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
