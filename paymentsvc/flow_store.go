package paymentsvc

import (
	"strings"
	"sync"

	"github.com/google/uuid"
)

type flowStore struct {
	mu    sync.Mutex
	items map[string]*pendingFlow
}

func (s *flowStore) put(flow *pendingFlow) string {
	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[id] = flow
	return id
}

func (s *flowStore) get(id string) *pendingFlow {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.items[strings.TrimSpace(id)]
}

func (s *flowStore) pop(id string) *pendingFlow {
	s.mu.Lock()
	defer s.mu.Unlock()
	id = strings.TrimSpace(id)
	flow := s.items[id]
	delete(s.items, id)
	return flow
}

func (s *flowStore) close() {
	s.mu.Lock()
	flows := make([]*pendingFlow, 0, len(s.items))
	for _, flow := range s.items {
		flows = append(flows, flow)
	}
	s.items = map[string]*pendingFlow{}
	s.mu.Unlock()
	for _, flow := range flows {
		flow.close()
	}
}
