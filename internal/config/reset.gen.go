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

// Reset resets the struct to its zero values.
func (s *ServerFlags) Reset() {
	if s == nil {
		return
	}

	s.ServerAddr = ""
	s.PostgresDSN = ""
	s.HashKey = ""
	s.ServerAuditFlags.Reset()
	s.ServerFileFlags.Reset()
}

// Reset resets the struct to its zero values.
func (s *ServerFileFlags) Reset() {
	if s == nil {
		return
	}

	s.FilePath = ""
	s.FileIsRestore = false
}

// Reset resets the struct to its zero values.
func (s *ServerAuditFlags) Reset() {
	if s == nil {
		return
	}

	s.AuditFilePath = ""
	s.AuditURL = ""
}
