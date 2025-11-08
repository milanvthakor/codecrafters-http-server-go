package server

import (
	"fmt"
	"io"
	"net"
	"os"

	"github.com/codecrafters-io/http-server-starter-go/app/config"
	"github.com/codecrafters-io/http-server-starter-go/app/handler"
	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

// Server represents the HTTP server
type Server struct {
	config *config.Config
	router *handler.Router
}

// NewServer creates a new HTTP server instance
func NewServer(cfg *config.Config) *Server {
	handlerConfig := &handler.Config{
		Directory: cfg.GetDirectory(),
	}

	return &Server{
		config: cfg,
		router: handler.NewRouter(handlerConfig),
	}
}

// Start starts the HTTP server and begins accepting connections
func (s *Server) Start() error {
	address := "0.0.0.0:" + s.config.Port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to bind to port %s: %w", s.config.Port, err)
	}
	defer listener.Close()

	fmt.Fprintf(os.Stdout, "Server listening on %s\n", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		parser := http.NewParser(conn)
		req, err := parser.ParseRequest()
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(os.Stderr, "Error parsing request: %v\n", err)
			}
			return
		}

		// Handle the request
		if err := s.router.HandleRequest(req, conn); err != nil {
			fmt.Fprintf(os.Stderr, "Error handling request: %v\n", err)
		}

		// Check if connection should be closed
		if s.router.ShouldCloseConnection(req) {
			conn.Close()
			return
		}
	}
}
