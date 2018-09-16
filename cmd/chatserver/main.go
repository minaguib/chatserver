package main

import (
	"flag"
	"github.com/minaguib/chatserver/internal/app/chatserver"
	"github.com/minaguib/chatserver/internal/app/chatserver/config"
)

func main() {

	portPtr := flag.String("port", config.ServerDefaultPort, "TCP port to listen on")
	flag.Parse()

	chatserver.RunServer(*portPtr)
}
