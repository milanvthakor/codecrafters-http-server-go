package http

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	CRLF = "\r\n"
)

// Request represents an HTTP request
type Request struct {
	Method        string
	RequestTarget string
	Version       string
	Headers       map[string]string
}

// Parser handles parsing of HTTP requests
type Parser struct {
	conn net.Conn
}

// NewParser creates a new request parser for a connection
func NewParser(conn net.Conn) *Parser {
	return &Parser{conn: conn}
}

// ParseRequest parses a complete HTTP request from the connection
func (p *Parser) ParseRequest() (*Request, error) {
	req := &Request{
		Headers: make(map[string]string),
	}

	// Parse request line
	if err := p.parseRequestLine(req); err != nil {
		return nil, err
	}

	// Parse headers
	if err := p.parseHeaders(req); err != nil {
		return nil, err
	}

	return req, nil
}

// ReadBody reads the request body based on Content-Length header
func (p *Parser) ReadBody(req *Request) ([]byte, error) {
	contentLengthStr, ok := req.Headers["Content-Length"]
	if !ok {
		return nil, errors.New("header 'Content-Length' is missing")
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Length: %w", err)
	}

	data := make([]byte, contentLength)
	if _, err := p.conn.Read(data); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	return data, nil
}

// parseRequestLine parses the HTTP request line (method, target, version)
func (p *Parser) parseRequestLine(req *Request) error {
	line, err := p.readUntilCRLF()
	if err != nil {
		return fmt.Errorf("failed to read request line: %w", err)
	}

	tokens := strings.Split(line, " ")
	if len(tokens) != 3 {
		return fmt.Errorf("invalid request line: expected 3 tokens, got %d", len(tokens))
	}

	req.Method = tokens[0]
	req.RequestTarget = tokens[1]
	req.Version = tokens[2]

	return nil
}

// parseHeaders parses HTTP headers until an empty line
func (p *Parser) parseHeaders(req *Request) error {
	for {
		line, err := p.readUntilCRLF()
		if err != nil {
			return fmt.Errorf("failed to read header: %w", err)
		}

		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header format: %s", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		req.Headers[key] = value
	}

	return nil
}

// readUntilCRLF reads from the connection until it finds a CRLF sequence
func (p *Parser) readUntilCRLF() (string, error) {
	p.conn.SetReadDeadline(time.Now().Add(time.Second))
	defer p.conn.SetReadDeadline(time.Time{})

	reader := bufio.NewReader(p.conn)
	var buf bytes.Buffer

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return buf.String(), io.EOF
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return buf.String(), nil
			}
			return "", err
		}

		buf.Write(line)
		result := buf.String()

		if len(result) >= 2 && result[len(result)-2:] == CRLF {
			return result[:len(result)-2], nil
		}
	}
}
