package main

import (
	"cmd/server"
	"cmd/stores"
)

func main() {
	store, err := stores.NewUserStore()
	if err != nil {
		println("Error initializing store:", err)
		return
	}
	err = server.Start(store)
	if err != nil {
		println("Error starting server:", err)
	}
}
