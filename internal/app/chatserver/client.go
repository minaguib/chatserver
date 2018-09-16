package chatserver

import (
	"bufio"
	"github.com/minaguib/chatserver/internal/app/chatserver/config"
	"golang.org/x/time/rate"
	"io"
	"net"
	"strings"
	"time"
)

type client struct {
	conn     net.Conn
	ip       string
	name     string
	scanner  *bufio.Scanner
	writeCH  chan (string)
	isClosed bool
	server   *chatServer
	limiter  *rate.Limiter // For handling incoming floods
}

func clientHandleNew(conn net.Conn, server *chatServer) {

	client := &client{
		conn:    conn,
		ip:      conn.RemoteAddr().String(),
		scanner: bufio.NewScanner(conn),
		writeCH: make(chan string, config.ClientOutputMaxBufferedMessages),
		server:  server,
		limiter: rate.NewLimiter(config.ClientInputMaxRatePerSec, config.ClientInputMaxRateBurst),
	}

	name, err := client.login()
	if err != nil {
		client.close("no-name")
		return
	}
	client.name = name

	go client.handleWrites()
	client.server.bus.announce(&busMessageClientConnected{client})

	client.doRead()

}

func (client *client) close(reason string) {
	if client.isClosed {
		return
	}
	client.isClosed = true
	close(client.writeCH)
	client.conn.Close()
	/* In case we were called from tryWrite which was called from the server wheel processor,
	write back on the bus asynchronously to avoid a potential deadlock */
	go func() {
		client.server.bus.announce(&busMessageClientDisconnected{client, reason})
	}()
}

func (client *client) tryWriteMessage(author string, message string) {
	line := time.Now().Format(config.ClientOutputTimeFormat) + ": [" + author + "] " + message + "\n"
	client.tryWrite(line)
}

func (client *client) tryWrite(line string) {
	if client.isClosed {
		return
	}
	if len(client.writeCH) < cap(client.writeCH) {
		client.writeCH <- line
		return
	}
	client.close("write-queue-overflow")
}

func (client *client) handleWrites() {
	for m := range client.writeCH {
		client.conn.SetWriteDeadline(time.Now().Add(config.ClientOutputWriteTimeoutSecs * time.Second))
		if _, err := io.WriteString(client.conn, m); err != nil {
			client.close("write-timeout")
			break
		}
	}
}

func (client *client) doRead() {
	warnIfOverflow := true
	for client.scanner.Scan() {
		text := strings.TrimSpace(client.scanner.Text())
		if text == "" {
			continue
		}
		if client.limiter.Allow() {
			client.server.bus.announce(&busMessageClientText{client, text})
			warnIfOverflow = true
		} else {
			if warnIfOverflow {
				client.tryWriteMessage("SYSTEM", "Message ["+text+"] onwards ignored for a while.  Slow down.")
			}
			warnIfOverflow = false
		}
	}
	client.close("read-eof")
}

func (client *client) login() (name string, err error) {

	client.conn.SetDeadline(time.Now().Add(config.ClientInputLoginTimeoutSecs * time.Second))
	defer client.conn.SetDeadline(time.Time{})

	_, err = io.WriteString(client.conn, "Welcome to the chat server\n")
	if err != nil {
		return name, err
	}

	for name == "" {
		if _, err := io.WriteString(client.conn, "Choose name: "); err != nil {
			break
		}
		if client.scanner.Scan() == false {
			break
		}
		name = strings.TrimSpace(client.scanner.Text())
		if name == "" {
			if _, err := io.WriteString(client.conn, "Invalid name.  Try again.\n"); err != nil {
				break
			}
		}
	}

	return name, err

}
