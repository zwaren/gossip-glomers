package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type OtherNode struct {
	ID string
}

type Server struct {
	node     *maelstrom.Node
	storage  *Storage
	topology []*OtherNode
}

func (s *Server) gossip(msg int, srcID string) {
	// fmt.Println("topology:", s.topology)
	for _, dest := range s.topology {
		if dest.ID != srcID {
			s.node.Send(dest.ID, map[string]any{
				"type":     "gossip",
				"message":  msg,
				"checkSum": s.storage.checkSum,
				"checkLen": len(s.storage.kv),
			})
		}
	}
}

func (s *Server) conflict(msgs []int, dstID string) {
	s.node.Send(dstID, map[string]any{
		"type":     "conflict",
		"messages": msgs,
		"checkSum": s.storage.checkSum,
		"checkLen": len(s.storage.kv),
	})
}

func (s *Server) checkConflict(checkSum int, checkLen int, srcID string) {
	if s.storage.checkSum != checkSum || len(s.storage.kv) != checkLen {
		msgs := s.storage.ReadAll()
		s.conflict(msgs, srcID)
	}
}

func (s *Server) paranoid(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, dst := range s.topology {
				msgs := s.storage.ReadAll()
				s.conflict(msgs, dst.ID)
			}
		case <-ctx.Done():
			return
		}
	}
}

// ----------- Public API -------------

func (s *Server) broadcastHandler(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	message := int(body["message"].(float64))
	s.storage.Add(message)
	s.gossip(message, msg.Src)

	return s.node.Reply(msg, map[string]any{
		"type": "broadcast_ok",
	})
}

func (s *Server) readHandler(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	return s.node.Reply(msg, map[string]any{
		"type":     "read_ok",
		"messages": s.storage.ReadAll(),
	})
}

func (s *Server) topologyHandler(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	// fmt.Println("topology-type:", reflect.TypeOf(body["topology"]))
	topology := body["topology"].(map[string]any)
	myTopology := topology[s.node.ID()].([]any)
	s.topology = make([]*OtherNode, 0, len(myTopology))
	for _, dest := range myTopology {
		destStr := dest.(string)
		s.topology = append(s.topology, &OtherNode{ID: destStr})
	}

	return s.node.Reply(msg, map[string]any{
		"type": "topology_ok",
	})
}

// ----------- Private API -------------

func (s *Server) gossipHandler(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	message := int(body["message"].(float64))
	if !s.storage.Add(message) {
		return nil
	}

	// checkSum := int(body["checkSum"].(float64))
	// checkLen := int(body["checkLen"].(float64))
	// s.checkConflict(checkSum, checkLen, msg.Src)
	s.gossip(message, msg.Src)
	return nil
}

func (s *Server) conflictHandler(msg maelstrom.Message) error {
	var body map[string]any
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	messages := body["messages"].([]any)
	s.storage.AddBatchAny(messages)

	checkSum := int(body["checkSum"].(float64))
	checkLen := int(body["checkLen"].(float64))
	s.checkConflict(checkSum, checkLen, msg.Src)
	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := &Server{
		node:    maelstrom.NewNode(),
		storage: NewStorage(),
	}

	// Public API
	s.node.Handle("broadcast", s.broadcastHandler)
	s.node.Handle("read", s.readHandler)
	s.node.Handle("topology", s.topologyHandler)

	// Private API
	s.node.Handle("gossip", s.gossipHandler)
	s.node.Handle("conflict", s.conflictHandler)

	go s.paranoid(ctx)

	if err := s.node.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
