package chatserver

/*
Bus provides a channel to send announcements
from clients to the main server wheel
*/

type bus chan busMessage

type busMessage interface {
}
type busMessageClientConnected struct {
	client *client
}
type busMessageClientDisconnected struct {
	client *client
	reason string
}
type busMessageClientText struct {
	client *client
	text   string
}

func (bus bus) announce(message busMessage) {
	bus <- message
}
