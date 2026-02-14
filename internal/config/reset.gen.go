package config

// Reset resets the struct to its zero values.
func (a *AgentFlags) Reset() {
	if a == nil {
		return
	}

	a.ServerAddr = ""
	a.HashKey = ""
	a.RateLimit = 0
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
