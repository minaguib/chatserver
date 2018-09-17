package chatserver

import (
	"log"
	"net"
	"os"

	"github.com/minaguib/chatserver/internal/app/chatserver/config"
)

type chatServer struct {
	port     string
	listener net.Listener
	bus      bus
	clients  map[*client]bool
	logger   *log.Logger
}

// RunServer is the entry point to run the Chat Server
// Does not return
func RunServer(port string) {
	server := &chatServer{
		port:    port,
		bus:     make(bus, config.ServerBusMaxBufferedMessages),
		clients: make(map[*client]bool),
		logger:  log.New(os.Stdout, "", log.LstdFlags),
	}
	server.startListener()
	server.logger.Println("Chat server now listening on port", server.port)
	go server.acceptNewConnections()
	server.wheel()
}

func (server *chatServer) wheel() {

	for m := range server.bus {
		switch m := m.(type) {
		case *busMessageClientConnected:
			server.clients[m.client] = true
			server.logger.Print("New client IP: ", m.client.ip, " name: ", m.client.name, " - ", len(server.clients), " client(s) connected")
			for peer := range server.clients {
				if m.client != peer {
					peer.tryWriteMessage("SYSTEM", m.client.name+" has joined the chat")
				}
			}
		case *busMessageClientDisconnected:
			if server.clients[m.client] {
				delete(server.clients, m.client)
				server.logger.Print("Goodbye client IP: ", m.client.ip, " name [", m.client.name, "] : ", m.reason, " - ", len(server.clients), " client(s) connected")
				for peer := range server.clients {
					peer.tryWriteMessage("SYSTEM", m.client.name+" has left the chat")
				}
			}
		case *busMessageClientText:
			for peer := range server.clients {
				if m.client != peer {
					peer.tryWriteMessage(m.client.name, m.text)
				}
			}
		default:
			server.logger.Print("Unknown bus message: ", m)
		}
	}

}

func (server *chatServer) acceptNewConnections() {
	for {
		if conn, err := server.listener.Accept(); err != nil {
			server.logger.Print("Failed to accept new connection: ", err)
		} else {
			go clientHandleNew(conn, server)
		}
	}

}

func (server *chatServer) startListener() {
	listener, err := net.Listen("tcp", ":"+server.port)
	if err != nil {
		panic(err)
	}
	server.listener = listener
}
