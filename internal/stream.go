package internal

import (
	"client-server-rabbitmq/api"
	"container/list"
	"encoding/json"
	"fmt"
	"os"
)

// Data represents the queue belonging to a client.
// It has a doubly linked list, with each key also
// referenced by a map. This results in all operations
// being O(1)
type Data struct {
	itemList *list.List
	// duplicate keys are kept in a list
	itemMap map[string][]*list.Element
}

func NewData() *Data {
	return &Data{
		list.New(),
		map[string][]*list.Element{},
	}
}

func (d *Data) AddItem(item api.Item) {
	e := d.itemList.PushBack(item)
	d.itemMap[item.Key] = append(d.itemMap[item.Key], e)
}

// Get all items for this key
func (d *Data) GetItem(key string) []api.Item {
	rv := []api.Item{}
	elems, ok := d.itemMap[key]
	if !ok {
		// for a removed key, return an invalid id
		return []api.Item{{key, -1}}
	}
	for _, e := range elems {
		rv = append(rv, e.Value.(api.Item))
	}
	return rv
}

func (d *Data) RemoveItem(key string) {
	if elems, ok := d.itemMap[key]; ok {
		for _, e := range elems {
			d.itemList.Remove(e)
		}
		delete(d.itemMap, key)
	}
}

// Get all items in the stream
func (d *Data) GetAllItems() []api.Item {
	rv := []api.Item{}
	for e := d.itemList.Front(); e != nil; e = e.Next() {
		rv = append(rv, e.Value.(api.Item))
	}
	return rv
}

func WriteDataToFile(output api.Output, f *os.File) error {
	data, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to serialize output: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	if _, err := f.Write([]byte("\n")); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}
