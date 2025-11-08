package handler

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"

	"octo-server/app/compression"
	"octo-server/app/http"
)

var (
	EchoEndpointRegex = regexp.MustCompile(`^/echo/(.+)$`)
	FileEndpointRegex = regexp.MustCompile(`^/files/(.+)$`)
)

// HandlerFunc is the type for HTTP handler functions
type HandlerFunc func(req *http.Request, writer *http.Writer, config *Config) error

// Config holds handler configuration
type Config struct {
	Directory string
}

// RootHandler handles the root endpoint
func RootHandler(req *http.Request, writer *http.Writer, config *Config) error {
	resp := &http.Response{
		StatusCode: 200,
		StatusText: http.StatusCodeToText(200),
		Headers:    make(map[string]string),
		Body:       nil,
	}
	return writer.WriteResponse(resp)
}

// NotFoundHandler handles 404 responses
func NotFoundHandler(req *http.Request, writer *http.Writer, config *Config) error {
	resp := &http.Response{
		StatusCode: 404,
		StatusText: http.StatusCodeToText(404),
		Headers:    make(map[string]string),
		Body:       nil,
	}
	return writer.WriteResponse(resp)
}

// BadRequestHandler handles 400 responses
func BadRequestHandler(req *http.Request, writer *http.Writer, config *Config) error {
	resp := &http.Response{
		StatusCode: 400,
		StatusText: http.StatusCodeToText(400),
		Headers:    make(map[string]string),
		Body:       nil,
	}
	return writer.WriteResponse(resp)
}

// InternalServerErrorHandler handles 500 responses
func InternalServerErrorHandler(req *http.Request, writer *http.Writer, config *Config) error {
	resp := &http.Response{
		StatusCode: 500,
		StatusText: http.StatusCodeToText(500),
		Headers:    make(map[string]string),
		Body:       nil,
	}
	return writer.WriteResponse(resp)
}

// EchoHandler handles the /echo/<str> endpoint
func EchoHandler(req *http.Request, writer *http.Writer, config *Config) error {
	matches := EchoEndpointRegex.FindStringSubmatch(req.RequestTarget)
	if len(matches) < 2 {
		return NotFoundHandler(req, writer, config)
	}

	str := matches[1]
	compressor := compression.NewCompressor()

	resp := &http.Response{
		StatusCode: 200,
		StatusText: http.StatusCodeToText(200),
		Headers:    make(map[string]string),
	}

	acceptEncoding := req.Headers["Accept-Encoding"]
	if compressor.SupportsGzip(acceptEncoding) {
		compressed, err := compressor.CompressGzip([]byte(str))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to compress data: %v\n", err)
			return InternalServerErrorHandler(req, writer, config)
		}

		resp.Headers["Content-Type"] = "text/plain"
		resp.Headers["Content-Encoding"] = "gzip"
		resp.Headers["Content-Length"] = fmt.Sprintf("%d", len(compressed))
		resp.Body = compressed
	} else {
		resp.Headers["Content-Type"] = "text/plain"
		resp.Headers["Content-Length"] = fmt.Sprintf("%d", len(str))
		resp.Body = []byte(str)
	}

	return writer.WriteResponse(resp)
}

// UserAgentHandler handles the /user-agent endpoint
func UserAgentHandler(req *http.Request, writer *http.Writer, config *Config) error {
	userAgent, ok := req.Headers["User-Agent"]
	if !ok {
		fmt.Fprintf(os.Stderr, "No 'User-Agent' header present!\n")
		os.Exit(1)
	}

	resp := &http.Response{
		StatusCode: 200,
		StatusText: http.StatusCodeToText(200),
		Headers: map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": fmt.Sprintf("%d", len(userAgent)),
		},
		Body: []byte(userAgent),
	}

	return writer.WriteResponse(resp)
}

// GetFileHandler handles GET /files/{filename} endpoint
func GetFileHandler(req *http.Request, writer *http.Writer, config *Config) error {
	if config.Directory == "" {
		fmt.Fprintf(os.Stderr, "Directory not configured\n")
		return InternalServerErrorHandler(req, writer, config)
	}

	matches := FileEndpointRegex.FindStringSubmatch(req.RequestTarget)
	if len(matches) < 2 || matches[1] == "" {
		return BadRequestHandler(req, writer, config)
	}

	filename := matches[1]
	filepath := config.Directory + "/" + filename

	file, err := os.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return NotFoundHandler(req, writer, config)
		}
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		return InternalServerErrorHandler(req, writer, config)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		return InternalServerErrorHandler(req, writer, config)
	}

	resp := &http.Response{
		StatusCode: 200,
		StatusText: http.StatusCodeToText(200),
		Headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(content)),
		},
		Body: content,
	}

	return writer.WriteResponse(resp)
}

// SaveFileHandler handles POST /files/{filename} endpoint
func SaveFileHandler(req *http.Request, writer *http.Writer, config *Config, parser *http.Parser) error {
	if config.Directory == "" {
		fmt.Fprintf(os.Stderr, "Directory not configured\n")
		return InternalServerErrorHandler(req, writer, config)
	}

	matches := FileEndpointRegex.FindStringSubmatch(req.RequestTarget)
	if len(matches) < 2 || matches[1] == "" {
		return BadRequestHandler(req, writer, config)
	}

	filename := matches[1]
	filepath := config.Directory + "/" + filename

	body, err := parser.ReadBody(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read request body: %v\n", err)
		return InternalServerErrorHandler(req, writer, config)
	}

	if err := os.WriteFile(filepath, body, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write file: %v\n", err)
		return InternalServerErrorHandler(req, writer, config)
	}

	resp := &http.Response{
		StatusCode: 201,
		StatusText: http.StatusCodeToText(201),
		Headers:    make(map[string]string),
		Body:       nil,
	}

	return writer.WriteResponse(resp)
}
