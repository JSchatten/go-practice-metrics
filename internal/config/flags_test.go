package config

import (
	"flag"
	"os"
	"testing"
	"time"
)

// clearFlags сбрасывает состояние флагов для тестов
func clearFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestInitAgentFlags_PriorityFlagOverDefault(t *testing.T) {
	clearFlags()

	// Убедимся, что env не заданы
	unsetEnv := func(key string) func() {
		if v, ok := os.LookupEnv(key); ok {
			os.Unsetenv(key)
			return func() { os.Setenv(key, v) }
		}
		return func() {}
	}
	defer unsetEnv("POLL_INTERVAL")()
	defer unsetEnv("REPORT_INTERVAL")()
	defer unsetEnv("ADDRESS")()

	// Только флаги
	os.Args = []string{"cmd", "-p", "7", "-r", "15", "-a", "flag-host:8080"}

	agentFlags, err := InitAgentFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agentFlags.PollInterval != 7*time.Second {
		t.Errorf("expected PollInterval=7s, got %v", agentFlags.PollInterval)
	}
	if agentFlags.ReportInterval != 15*time.Second {
		t.Errorf("expected ReportInterval=15s, got %v", agentFlags.ReportInterval)
	}
	if agentFlags.ServerAddr != "flag-host:8080" {
		t.Errorf("expected ServerAddr=flag-host:8080, got %s", agentFlags.ServerAddr)
	}
}

func TestInitAgentFlags_DefaultValues(t *testing.T) {
	clearFlags()

	// Убираем все переменные окружения
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
	os.Unsetenv("ADDRESS")

	// Без флагов — должны использоваться значения по умолчанию
	os.Args = []string{"cmd"}

	agentFlags, err := InitAgentFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agentFlags.PollInterval != 2*time.Second {
		t.Errorf("expected default PollInterval=2s, got %v", agentFlags.PollInterval)
	}
	if agentFlags.ReportInterval != 10*time.Second {
		t.Errorf("expected default ReportInterval=10s, got %v", agentFlags.ReportInterval)
	}
	if agentFlags.ServerAddr != "localhost:8080" {
		t.Errorf("expected default ServerAddr=localhost:8080, got %s", agentFlags.ServerAddr)
	}
}

func TestInitAgentFlags_InvalidPollInterval(t *testing.T) {
	clearFlags()

	os.Args = []string{"cmd", "-p", "-1"}
	_, err := InitAgentFlags()
	if err != ErrInvalidPollInterval {
		t.Errorf("expected ErrInvalidPollInterval, got %v", err)
	}
}

func TestInitAgentFlags_InvalidReportInterval(t *testing.T) {
	clearFlags()

	os.Args = []string{"cmd", "-r", "-5"}
	_, err := InitAgentFlags()
	if err != ErrInvalidReportInterval {
		t.Errorf("expected ErrInvalidReportInterval, got %v", err)
	}
}

func TestInitServerFlags_PriorityFlagOverDefault(t *testing.T) {
	clearFlags()

	os.Unsetenv("ADDRESS")
	os.Args = []string{"cmd", "-a", "flag-server:8080"}

	serverFlags, err := InitServerFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if serverFlags.ServerAddr != "flag-server:8080" {
		t.Errorf("expected ServerAddr=flag-server:8080, got %s", serverFlags.ServerAddr)
	}
}

func TestInitServerFlags_DefaultValue(t *testing.T) {
	clearFlags()

	os.Unsetenv("ADDRESS")
	os.Args = []string{"cmd"}

	serverFlags, err := InitServerFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if serverFlags.ServerAddr != "localhost:8080" {
		t.Errorf("expected default ServerAddr=localhost:8080, got %s", serverFlags.ServerAddr)
	}
}

func TestInitFlags_UnknownArgs(t *testing.T) {
	clearFlags()

	os.Args = []string{"cmd", "extra-arg"}
	_, err := InitAgentFlags()
	if err == nil {
		t.Fatal("expected error for unknown args, got nil")
	}
	if err.Error() != "error: unknown flags: [extra-arg]" {
		t.Errorf("unexpected error message: %v", err)
	}
}
