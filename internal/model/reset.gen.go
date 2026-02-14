package model

// Reset resets the struct to its zero values.
func (m *Metrics) Reset() {
	if m == nil {
		return
	}

	m.ID = ""
	m.MType = ""
	if m.Delta != nil {
		*m.Delta = 0
	}
	if m.Value != nil {
		*m.Value = 0.0
	}
	m.Hash = ""
}
