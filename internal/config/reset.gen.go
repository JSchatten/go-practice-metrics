package config

// Reset resets the struct to its zero values.
func (s *AgentFlags) Reset() {
	if s == nil {
		return
	}
	s.ServerAddr = ""
	s.HashKey = ""
	s.RateLimit = 0

}
