package main

import (
	"client-server-rabbitmq/internal"
	"log"
	"os"
)

func main() {

	cfg, err := internal.GetConfig()
	if err != nil {
		log.Panicf("failed to get config: %v", err)
	}

	f, err := os.OpenFile(cfg.SERVER_OUTPUT_FILENAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panicf("failed to open file: %v", err)
	}

	serv := internal.NewServer(cfg.USERNAME, cfg.PASSWORD, cfg.QUEUE_NAME, cfg.RABBIT_MQ_PORT, f)
	defer serv.SafeClose()
	serv.Start()

	ch := make(chan struct{})
	<-ch
}
