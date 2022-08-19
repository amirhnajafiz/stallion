package stallion

import (
	"fmt"
	"net"

	"github.com/amirhnajafiz/stallion/internal"
	"go.uber.org/zap"
)

type Server interface {
	// Handle method generates a new worker for clients.
	Handle(conn net.Conn)
}

// NewServer creates a new broker server on given port.
func NewServer(port string) error {
	// creating a new server
	serve := internal.NewServer()

	// listen over a port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	zap.L().Info("start broker server", zap.String("port", port))

	// handling our clients
	for {
		if conn, er := listener.Accept(); er == nil {
			serve.Handle(conn)
		} else {
			zap.L().Error("error in client accept", zap.Error(er))
		}
	}
}
