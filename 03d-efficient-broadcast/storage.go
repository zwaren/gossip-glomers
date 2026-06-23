package main

import "sync"

type Storage struct {
	mu       sync.RWMutex
	kv       map[int]struct{}
	checkSum int
}

func NewStorage() *Storage {
	return &Storage{kv: make(map[int]struct{})}
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
		s.checkSum ^= msg
		return true
	}
	return false
}

func (s *Storage) AddBatch(msgs []int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := false
	for _, msg := range msgs {
		if _, ok := s.kv[msg]; !ok {
			s.kv[msg] = struct{}{}
			s.checkSum ^= msg
			updated = true
		}
	}
	return updated
}

func (s *Storage) AddBatchAny(msgs []any) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	updated := false
	for _, msg := range msgs {
		msgInt := int(msg.(float64))
		if _, ok := s.kv[msgInt]; !ok {
			s.kv[msgInt] = struct{}{}
			s.checkSum ^= msgInt
			updated = true
		}
	}
	return updated
}
