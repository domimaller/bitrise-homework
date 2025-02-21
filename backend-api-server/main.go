package main

import "backend-api-server/server"

func main() {
	server := server.New(server.NewConfig())
	server.Run()
}
