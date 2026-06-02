package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Storage struct {
	mu sync.RWMutex
	kv map[int]struct{}
}

func (s *Storage) ReadAll() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]int, 0, len(s.kv))
	for k := range s.kv {
		keys = append(keys, k)
	}
	return keys
}

func (s *Storage) Add(msg int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.kv[msg]; !ok {
		s.kv[msg] = struct{}{}
		return true
	}
	return false
}

var server *maelstrom.Node
var storage *Storage
var topologyMap map[string]any

func gossip(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	message := int(body["message"].(float64))
	if !storage.Add(message) {
		return nil
	}
	dests := topologyMap[server.ID()]
	for _, dest := range dests.([]any) {
		dst := dest.(string)
		if dst != msg.Src {
			server.Send(dst, map[string]any{
				"type":    "gossip",
				"message": message,
			})
		}
	}
	return nil
}

func broadcast(msg maelstrom.Message) error {
	if err := gossip(msg); err != nil {
		return err
	}
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
		"messages": storage.ReadAll(),
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
	storage = &Storage{kv: make(map[int]struct{})}

	server.Handle("broadcast", broadcast)
	server.Handle("read", read)
	server.Handle("topology", topology)
	server.Handle("gossip", gossip)

	if err := server.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
