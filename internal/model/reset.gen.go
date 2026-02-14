package model

// Reset resets the struct to its zero values.
func (s *Metrics) Reset() {
	if s == nil {
		return
	}

	s.ID = ""
	s.MType = ""
	if s.Delta != nil {
		*s.Delta = 0
	}
	if s.Value != nil {
		*s.Value = 0.0
	}
	s.Hash = ""
}
