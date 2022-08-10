package stallion

import (
	"fmt"
	"log"
	"net"

	"github.com/amirhnajafiz/stallion/internal"
)

// NewServer creates a new broker server on given port.
func NewServer(port string) error {
	// channels for public and status messages
	channel := make(chan internal.Message)
	sub := make(chan internal.SubscribeChannel)
	unsub := make(chan internal.UnsubscribeChannel)
	ter := make(chan int)

	// creating a new server
	serve := internal.NewServer(channel, sub, unsub, ter)

	// listen over a port
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	log.Printf("start broker server on %s ...\n", port)

	// handling our clients
	for {
		conn, _ := listener.Accept()

		serve.Handle(conn, channel, sub, unsub, ter)
	}
}
