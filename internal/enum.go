package internal

// constant values for message types
const (
	Text int = iota + 1
	Subscribe
	Unsubscribe
)

const (
	SubStatus int = iota + 1
	UnsubStatus
	TerminateStatus
)

const (
	DummyMessage = "hello world"
)
