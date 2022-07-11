package main

import (
	"bufio"
	"client-server-rabbitmq/api"
	"client-server-rabbitmq/internal"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testSpec struct {
	req              api.Request
	expectedResponse api.Output
}

var caps = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var TEST_OUTPUT_FILE = "test_output.json"

func TestClientServer(t *testing.T) {

	rand.Seed(time.Now().UnixNano())

	cfg, err := internal.GetConfig()
	if err != nil {
		log.Panicf("failed to get config: %v", err)
	}

	// remove the test file if it exists already
	if err := os.Remove(TEST_OUTPUT_FILE); err != nil {
		log.Panicf("failed to remove existing test file: %v", err)
	}
	f, err := os.OpenFile(TEST_OUTPUT_FILE, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panicf("failed to open test file: %v", err)
	}

	server := internal.NewServer(cfg.USERNAME, cfg.PASSWORD, cfg.QUEUE_NAME, cfg.RABBIT_MQ_PORT, f)
	defer server.SafeClose()
	server.Start()

	numClients := 5
	numRequestsPerClient := 50
	testSpecs := map[string][]testSpec{}
	for i := 0; i < numClients; i++ {
		clientId := uuid.New().String()
		testSpecs[clientId] = generateRequests(clientId, numRequestsPerClient)
	}

	for c, tss := range testSpecs {
		c, tss := c, tss
		go func() {
			for _, ts := range tss {
				server.Qs.PublishRequest(ts.req, c)
			}
		}()
	}

	// check the output after a delay
	time.Sleep(10 * time.Second)

	testOutputs := map[string][]api.Output{}
	file, err := os.Open(TEST_OUTPUT_FILE)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		o := api.Output{}
		err = json.Unmarshal(scanner.Bytes(), &o)
		if err != nil {
			log.Printf("failed to parse file output: %v", err)
		} else {
			testOutputs[o.ClientId] = append(testOutputs[o.ClientId], o)
		}
	}

	for client, v := range testSpecs {
		expectedOutput := []api.Output{}
		for _, ts := range v {
			if ts.req.Operation == "GetItem" || ts.req.Operation == "GetAllItems" {
				expectedOutput = append(expectedOutput, ts.expectedResponse)
			}
		}
		require.Equal(t, expectedOutput, testOutputs[client])
	}
}

func generateRequests(clientId string, numRequests int) []testSpec {

	rv := []testSpec{}
	curQueue := []api.Item{}
	for i := 1; i <= numRequests; i++ {
		op := rand.Intn(len(api.Operations))
		req := api.Request{RequestId: int64(i), ClientId: clientId, Operation: api.Operations[op]}
		resp := api.Output{RequestId: int64(i), ClientId: clientId, Operation: api.Operations[op], Items: []api.Item{}}

		switch req.Operation {

		case "AddItem":
			item := api.Item{Key: string(caps[rand.Intn(len(caps))]), Id: req.RequestId}
			req.Items = []api.Item{item}
			curQueue = append(curQueue, item)

		case "RemoveItem":
			if len(curQueue) > 0 {
				idx := rand.Intn(len(curQueue))
				req.Items = []api.Item{curQueue[idx]}
				indicesToRemove := map[int]struct{}{}
				for i, item := range curQueue {
					if item.Key == curQueue[idx].Key {
						indicesToRemove[i] = struct{}{}
					}
				}
				curQueue2 := []api.Item{}
				for i, item := range curQueue {
					if _, ok := indicesToRemove[i]; !ok {
						curQueue2 = append(curQueue2, item)
					}
				}
				curQueue = curQueue2
			}

		case "GetItem":
			if len(curQueue) > 0 {
				idx := rand.Intn(len(curQueue))
				req.Items = []api.Item{curQueue[idx]}
				for _, i := range curQueue {
					if i.Key == curQueue[idx].Key {
						resp.Items = append(resp.Items, i)
					}
				}
			}

		case "GetAllItems":
			resp.Items = append(resp.Items, curQueue...)
		}
		rv = append(rv, testSpec{req: req, expectedResponse: resp})
	}
	return rv

}
