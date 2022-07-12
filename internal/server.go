package internal

import (
	"client-server-rabbitmq/api"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

// Server has an id,
// a data stream for each client,
// a file to write output to, and the queue state
type Server struct {
	id           string
	rw           sync.RWMutex
	clientStream map[string]struct {
		c chan api.Request
		*Data
	}
	outFile    *os.File
	outChannel chan api.Output
	Qs         *QueueState
}

func NewServer(username, password, queueName string, port int, outFile *os.File) *Server {
	serv := &Server{
		id: uuid.New().String(),
		rw: sync.RWMutex{},
		clientStream: map[string]struct {
			c chan api.Request
			*Data
		}{},
		outFile:    outFile,
		outChannel: make(chan api.Output),
	}

	// start go routine for file writer
	go serv.writeDataToFile()

	qs, err := InitQueue(username, password, queueName, port)
	if err != nil {
		log.Panicf("failed to initialize queue: %v", err)
	}
	serv.Qs = qs

	return serv
}

// When the server starts consuming the messages from the queue,
// it parses each request and routes it to the appropriate client stream
func (serv *Server) Start() {

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		serv.SafeClose()
		os.Exit(1)
	}()

	msgs, err := serv.Qs.Channel.Consume(serv.Qs.Queue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Panicf("failed to consume from queue: %v", err)
	}
	go func() {
		for d := range msgs {
			serv.parseAndRouteRequest(d.Body)
		}
	}()

	log.Printf("Started server %s. [*] Waiting for messages. To exit press CTRL+C", serv.id)
}

func (serv *Server) SafeClose() {
	serv.Qs.SafeClose()
	serv.outFile.Close()
}

// processRequest parses an incoming request and routes it to
// the stream belonging to the client
func (serv *Server) parseAndRouteRequest(reqBody []byte) {

	req := api.Request{}
	err := json.Unmarshal(reqBody, &req)
	if err != nil {
		log.Panicf("failed to parse request: %v", err)
	}
	serv.rw.RLock()
	cs, ok := serv.clientStream[req.ClientId]
	serv.rw.RUnlock()
	if !ok {
		cs = struct {
			c chan api.Request
			*Data
		}{
			make(chan api.Request),
			NewData(),
		}
		serv.rw.Lock()
		serv.clientStream[req.ClientId] = cs
		serv.rw.Unlock()
		go serv.processRequest(cs.c)
	}
	cs.c <- req
}

func (serv *Server) processRequest(c chan api.Request) {
	for req := range c {
		serv.rw.RLock()
		d := serv.clientStream[req.ClientId].Data
		serv.rw.RUnlock()
		switch req.Operation {
		case "AddItem":
			if len(req.Items) > 0 {
				d.AddItem(req.Items[0])
			}
		case "RemoveItem":
			if len(req.Items) > 0 {
				d.RemoveItem(req.Items[0].Key)
			}
		case "GetItem":
			items := []api.Item{}
			if len(req.Items) > 0 {
				items = d.GetItem(req.Items[0].Key)
			}
			output := api.Output{
				RequestId: req.RequestId,
				ClientId:  req.ClientId,
				Operation: req.Operation,
				Items:     items,
			}
			serv.outChannel <- output
		case "GetAllItems":
			output := api.Output{
				RequestId: req.RequestId,
				ClientId:  req.ClientId,
				Operation: req.Operation,
				Items:     d.GetAllItems(),
			}
			serv.outChannel <- output
		default:
			log.Panicf("unknown operation: %s", req.Operation)
		}
	}
}

func (serv *Server) writeDataToFile() {
	for output := range serv.outChannel {
		data, err := json.Marshal(output)
		if err != nil {
			log.Fatalf("failed to serialize output: %v", err)
		}
		if _, err := serv.outFile.Write(data); err != nil {
			log.Fatalf("failed to write to file: %v", err)
		}
		if _, err := serv.outFile.Write([]byte("\n")); err != nil {
			log.Fatalf("failed to write to file: %v", err)
		}
	}
}
