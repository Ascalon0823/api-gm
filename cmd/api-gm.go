package main

import (
	"cmd/server"
)

func main() {
	if err := server.Start(); err != nil {
		println("Error starting server:", err)
	}
}
