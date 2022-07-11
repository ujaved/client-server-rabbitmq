package main

import (
	"client-server-rabbitmq/internal"
)

func main() {

	client := internal.NewClient()
	defer client.SafeClose()
	client.Start()
}
