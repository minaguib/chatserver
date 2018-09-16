package main

import (
	"flag"
	"github.com/minaguib/chatserver/internal/app/chatserver"
)

func main() {

	portPtr := flag.String("port", "2234", "TCP port to listen on")
	flag.Parse()

	chatserver.RunServer(*portPtr)
}
