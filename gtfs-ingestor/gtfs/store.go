package gtfs

import "sync"

type Store struct {
	Arrivals map[string]*ArrivalUpdate
	mu       sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		Arrivals: make(map[string]*ArrivalUpdate),
	}
}

func (s *Store) UpdateArrival(update *ArrivalUpdate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Arrivals[update.StopID] = update
}

func (s *Store) GetArrival(stopID string) (*ArrivalUpdate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	update, exists := s.Arrivals[stopID]
	return update, exists
}
