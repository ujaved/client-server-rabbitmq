package api

// Item is a key and an Id, which is the request Id
type Item struct {
	Key string `json:"key"`
	Id  int64  `json:"id"`
}

var Operations = []string{"AddItem", "RemoveItem", "GetItem", "GetAllItems"}

type Request struct {
	RequestId int64  `json:"id,omitempty"`
	ClientId  string `json:"clientId" validate:"required"`
	Operation string `json:"operation" validate:"required"`
	Items     []Item `json:"items"`
}

type Output Request
