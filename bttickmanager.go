package malgova

type btTickManager struct {
	observerAlgoIDs []string
}

func (s *btTickManager) addObserver(algoID string) {
	if s.observerAlgoIDs == nil {
		s.observerAlgoIDs = make([]string, 0)
	}
	s.observerAlgoIDs = append(s.observerAlgoIDs, algoID)
}
