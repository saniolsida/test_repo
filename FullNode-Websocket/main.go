package main

import (
	"signature-server/server"
	"signature-server/subscriber"
)

func main() {
	go subscriber.StartFullNodeSubscription()
	server.StartWebSocketServer()
}
