package internal

import (
	"bufio"
	"client-server-rabbitmq/api"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
)

// Client has an id, a file where it reads requests from
//, and a reference to the queue state
type Client struct {
	id            string
	inputFilename string
	qs            *QueueState
}

// Initiate a client
func NewClient() *Client {
	cfg, err := GetConfig()
	if err != nil {
		log.Panicf("failed to get config: %v", err)
	}
	client := &Client{
		id:            uuid.New().String(),
		inputFilename: cfg.CLIENT_INPUT_FILENAME,
	}

	qs, err := InitQueue(cfg.USERNAME, cfg.PASSWORD, cfg.QUEUE_NAME, cfg.RABBIT_MQ_PORT)
	if err != nil {
		log.Panicf("failed to initialize queue: %v", err)
	}
	client.qs = qs

	return client
}

// Starting a client reads requests from the configured input file
// line by line, increments and sets a request Id and publishes the
// request to the queue. Then it waits for a request to appear on
// stdin to be published
func (client *Client) Start() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		client.SafeClose()
		os.Exit(1)
	}()

	requestId := 1
	fileContents, err := os.ReadFile(client.inputFilename)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			log.Printf("No input file exists at location: %s", client.inputFilename)
		} else {
			log.Panicf("failed to read input file: %v", err)
		}
	} else {
		reqs := []api.Request{}
		err = json.Unmarshal(fileContents, &reqs)
		if err != nil {
			log.Panicf("failed to parse input file: %v", err)
		}
		log.Printf("number of requests parsed from file: %d", len(reqs))

		for _, req := range reqs {
			req.RequestId = int64(requestId)
			err = client.qs.PublishRequest(req, client.id)
			if err != nil {
				log.Panicf("failed to publish request: %v", err)
			}
			requestId++
		}
	}

	// read requests from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		req := api.Request{}
		err = json.Unmarshal(scanner.Bytes(), &req)
		if err != nil {
			log.Printf("failed to parse request: %v", err)
		} else {
			req.RequestId = int64(requestId)
			err = client.qs.PublishRequest(req, client.id)
			if err != nil {
				log.Panicf("failed to publish request: %v", err)
			}
			requestId++
		}
	}
}

func (client *Client) SafeClose() {
	client.qs.SafeClose()
}
