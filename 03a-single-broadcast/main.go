package main

import (
	"encoding/json"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

var server *maelstrom.Node
var messages []float64
var topologyMap map[string]any

func broadcast(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	messages = append(messages, body["message"].(float64))

	return server.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
}

func read(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	return server.Reply(msg, map[string]any{
		"type":     "read_ok",
		"messages": messages,
	})
}

func topology(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	// fmt.Println("topology-type:", reflect.TypeOf(body["topology"]))
	topologyMap = body["topology"].(map[string]any)

	return server.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}

func main() {
	server = maelstrom.NewNode()

	server.Handle("broadcast", broadcast)
	server.Handle("read", read)
	server.Handle("topology", topology)

	if err := server.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
