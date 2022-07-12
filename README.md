Client-Server Application With External Queue
----------------------------------------------

This is a multi-threaded client-server application where the server implements a queue for each client. Each item in the queue
is a key-value pair, where the key is a string and the value is the integer Id of the client's request.

The server's queue in memory is implemented as a doubly-linked list where each key is also mapped to its (list of) nodes in the
list. Duplicate keys are allowed. The use of a doubly-linked list along with a map results in all CRUD operations have runtime O(1).

Launch Queue
-------------
Run the make target `make build-env` to launch the rabbotmq docker container. This step is a pre-requisite to launching the server and the client.

Launch Server
--------------
Run the make target `make server` to launch the server. The RabbitMQ queue parameters are read from environment variables. The server writes its
output to a file specified in the environment variable `SERVER_OUTPUT_FILENAME`. 

Launch Client
---------------
Run the make target `make client` to launch the client. The RabbitMQ queue parameters are read from environment variables. 
The client reads an array of json requests from a file configured in the env variable `CLIENT_INPUT_FILENAME`. After all requests are read from the
input file, further requests can be read from std input. Each request is of the form:
```
{
  "Operation":"AddItem",
  "Items": [
             {
                "Key": "A"
             }
           ]
}
```
where there are four operations: 1) `AddItem`, 2) `RemoveItem`, 3) `GetItem`, 4) `GetAllItems`.
For `AddItem`, `RemoveItem`, and `GetItem`, a single key is specified in the `Items` array.
For `GetAllItems`, `Items` is omitted.

Each client reads from the same input file. The env varibale `CLIENT_INPUT_FILENAME` can be updated 
to point to a different output for each new client.

Output
--------
Each output object has the same structure as the request object, with the additional fields
of `requestId` and `clientId`. Furthermore, every item has a value (`Id`) that is the same as the `requestId` that produced it. 
For a non-existent key, the item id is set to -1. 


Testing
--------------
Run the make target `make test` to run tests. This includes an integration test
`TestClientServer` that runs 5 concurrent clients, each of whom sends 100 random 
requests to the queue. (These parameters can be changed in the test file `cmd/tests/integration_test.go`)
After a delay, the test reads the server output file and checks that each output object is correct.  
