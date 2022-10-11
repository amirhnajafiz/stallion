package internal

import (
	"fmt"
	"net"
	"strings"
	"time"
)

// worker handles a single client that wants
// to either subscribe or publish.
type worker struct {
	// each worker has its unique id
	id int

	// authentication fields
	// user of stallion client
	user string
	// pass of stallion client
	pass string

	// for network socket handling
	network network

	// send channel is used for getting data from broker (its private between broker and worker)
	sendChannel chan message

	// receive channel is used for sending data to broker (its public)
	receiveChannel chan message

	// subscribeChannel is a public channel for subscribing workers over a topic
	subscribeChannel chan subscribeChannel

	// unsubscribeChannel is a public channel for unsubscribing workers from a topic
	unsubscribeChannel chan unsubscribeChannel

	// terminateChannel create a channel for dead workers
	terminateChannel chan int
}

// newWorker generates a new worker.
func newWorker(
	id int,
	user string,
	pass string,
	conn net.Conn,
	sen, rec chan message,
	sub chan subscribeChannel,
	unsub chan unsubscribeChannel,
	ter chan int,
) *worker {
	return &worker{
		id:   id,
		user: user,
		pass: pass,
		network: network{
			connection: conn,
		},
		sendChannel:        sen,
		receiveChannel:     rec,
		subscribeChannel:   sub,
		unsubscribeChannel: unsub,
		terminateChannel:   ter,
	}
}

// start will start our worker.
func (w *worker) start() {
	// closing channel after we are done
	defer close(w.sendChannel)

	// check the ping pong connection
	if err := w.pong(); err != nil {
		w.terminateChannel <- w.id

		return
	}

	// start for input data
	go w.arrival()

	// wait for any interrupt in send channel
	for {
		data := <-w.sendChannel

		w.transfer(data)
	}
}

// get client ping message.
func (w *worker) pong() error {
	// creating a buffer
	var buffer = make([]byte, 2048)

	// read data from network
	tmp, er := w.network.get(buffer)
	if er != nil {
		return fmt.Errorf("client failed to ping: %w", er)
	}

	// get user request
	request, _ := decodeMessage(tmp)
	data := strings.Split(string(request.Data), ":")

	// check auth
	if w.user == data[0] && w.pass == data[1] {
		// send pong response
		if err := w.network.send(encodeMessage(newMessage(PongMessage, "", nil))); err != nil {
			return fmt.Errorf("failed to pong client: %w", err)
		}
	}

	// return sabotage message
	if err := w.network.send(encodeMessage(newMessage(Imposter, "", nil))); err != nil {
		return fmt.Errorf("failed to pong client: %w", err)
	}

	return nil
}

// transfer will send a data byte through handler.
func (w *worker) transfer(data message) {
	err := w.network.send(encodeMessage(data))
	if err != nil {
		logError("failed to send data", err)
	}

	time.Sleep(1 * time.Millisecond)
}

// arrival will check for input data from client.
func (w *worker) arrival() {
	var buffer = make([]byte, 2048)

	for {
		tmp, err := w.network.get(buffer)
		if err != nil {
			logError("failed to read data", err)

			break
		}

		m, er := decodeMessage(tmp)
		if er != nil {
			logError("parse error", er)

			continue
		}

		// handle input messages
		w.handle(m)
	}

	// announcing that the worker is done
	w.terminateChannel <- w.id
}

// handle method responses to given messages.
func (w *worker) handle(m *message) {
	switch m.Type {
	case Text:
		// passing data to broker channel
		w.receiveChannel <- *m
	case Subscribe:
		// passing subscribe message
		w.subscribeChannel <- subscribeChannel{
			id:      w.id,
			topic:   m.Topic,
			channel: w.sendChannel,
		}
	case Unsubscribe:
		// passing unsubscribe message
		w.unsubscribeChannel <- unsubscribeChannel{
			id:    w.id,
			topic: m.Topic,
		}
	}
}
