package handler

import (
	"net"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

// Router handles HTTP request routing
type Router struct {
	config *Config
}

// NewRouter creates a new router with the given configuration
func NewRouter(config *Config) *Router {
	return &Router{
		config: config,
	}
}

// HandleRequest routes an HTTP request to the appropriate handler
func (r *Router) HandleRequest(req *http.Request, conn net.Conn) error {
	writer := http.NewWriter(conn)
	parser := http.NewParser(conn)

	var handler HandlerFunc

	switch {
	case req.RequestTarget == "/":
		handler = RootHandler

	case req.RequestTarget == "/user-agent":
		handler = UserAgentHandler

	case EchoEndpointRegex.MatchString(req.RequestTarget):
		handler = EchoHandler

	case FileEndpointRegex.MatchString(req.RequestTarget):
		switch req.Method {
		case "GET":
			handler = GetFileHandler
		case "POST":
			// POST handler needs parser for reading body
			return r.handlePostFile(req, writer, parser)
		default:
			handler = NotFoundHandler
		}

	default:
		handler = NotFoundHandler
	}

	return handler(req, writer, r.config)
}

// handlePostFile handles POST requests to /files/{filename}
func (r *Router) handlePostFile(req *http.Request, writer *http.Writer, parser *http.Parser) error {
	return SaveFileHandler(req, writer, r.config, parser)
}

// ShouldCloseConnection checks if the connection should be closed based on request headers
func (r *Router) ShouldCloseConnection(req *http.Request) bool {
	connection, ok := req.Headers["Connection"]
	return ok && connection == "close"
}
