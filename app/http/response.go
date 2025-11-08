package http

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Response represents an HTTP response
type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

// Writer handles writing HTTP responses
type Writer struct {
	conn net.Conn
}

// NewWriter creates a new response writer for a connection
func NewWriter(conn net.Conn) *Writer {
	return &Writer{conn: conn}
}

// WriteResponse writes a complete HTTP response to the connection
func (w *Writer) WriteResponse(resp *Response) error {
	// Build status line
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s%s", resp.StatusCode, resp.StatusText, CRLF)

	// Build headers
	var headers strings.Builder
	for key, value := range resp.Headers {
		headers.WriteString(fmt.Sprintf("%s: %s%s", key, value, CRLF))
	}
	headers.WriteString(CRLF)

	// Combine all parts
	response := statusLine + headers.String()
	if resp.Body != nil {
		responseBytes := []byte(response)
		responseBytes = append(responseBytes, resp.Body...)
		response = string(responseBytes)
	}

	// Write to connection
	if _, err := w.conn.Write([]byte(response)); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing response: %v\n", err)
		return err
	}

	return nil
}

// StatusCodeToText converts HTTP status code to status text
func StatusCodeToText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 400:
		return "Bad Request"
	case 404:
		return "Not Found"
	case 500:
		return "Internal Server Error"
	default:
		return "Unknown"
	}
}
