package internal

import (
	"net"
)

// server is our broker service.
type server struct {
	prefix int
	broker *broker
}

// NewServer returns a new broker server.
func NewServer(public chan Message, sub chan SubscribeChannel, unsub chan UnsubscribeChannel, ter chan int) *server {
	s := &server{
		prefix: 101,
	}

	// setting up the server broker and starting it
	s.broker = newBroker(public, sub, unsub, ter)
	go s.broker.start()

	return s
}

// Handle will handle the clients.
func (s *server) Handle(conn net.Conn, public chan Message, sub chan SubscribeChannel, unsub chan UnsubscribeChannel, ter chan int) {
	w := newWorker(s.prefix, conn, make(chan Message), public, sub, unsub, ter)

	s.prefix++

	go w.start()
}
