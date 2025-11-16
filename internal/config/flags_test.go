// internal/config/flags_test.go
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

// helper: временно устанавливает env и возвращает функцию для восстановления
func withEnv(key, value string) func() {
	oldValue, exists := os.LookupEnv(key)
	os.Setenv(key, value)
	return func() {
		if exists {
			os.Setenv(key, oldValue)
		} else {
			os.Unsetenv(key)
		}
	}
}

func TestInitAgentFlags_PriorityFlagOverDefault(t *testing.T) {
	clearFlags()
	defer withEnv("POLL_INTERVAL", "")()
	defer withEnv("REPORT_INTERVAL", "")()
	defer withEnv("ADDRESS", "")()

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
	defer withEnv("POLL_INTERVAL", "")()
	defer withEnv("REPORT_INTERVAL", "")()
	// defer withEnv("ADDRESS", "")()

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
	// if agentFlags.ServerAddr != "localhost:8080" {
	// 	t.Errorf("expected default ServerAddr=localhost:8080, got %s", agentFlags.ServerAddr)
	// }
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
	defer withEnv("ADDRESS", "")()

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
	// не нужно, т.к. по умолчанию localhost:8080
	// defer withEnv("ADDRESS", "")()
	defer withEnv("FILE_STORAGE_PATH", "")()
	defer withEnv("STORE_INTERVAL", "")()
	defer withEnv("RESTORE", "")()
	defer withEnv("DATABASE_DSN", "")()

	os.Args = []string{"cmd"}

	serverFlags, err := InitServerFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if serverFlags.ServerAddr != "localhost:8080" {
		t.Errorf("expected default ServerAddr=localhost:8080, got %s", serverFlags.ServerAddr)
	}
	if serverFlags.PostgresDSN != "" {
		t.Errorf("expected empty DSN by default, got %s", serverFlags.PostgresDSN)
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

func TestInitServerFlags_DSNFlagEmpty_Error(t *testing.T) {
	clearFlags()
	defer withEnv("DATABASE_DSN", "")()

	os.Args = []string{"cmd", "-d", ""}

	_, err := InitServerFlags()
	if err != ErrInvalidDSN {
		t.Fatalf("expected ErrInvalidDSN, got %v", err)
	}
}

func TestInitServerFlags_DSNEnvEmpty_Error(t *testing.T) {
	clearFlags()
	defer withEnv("DATABASE_DSN", "")()

	withEnv("DATABASE_DSN", "")()

	_, err := InitServerFlags()
	if err != ErrInvalidDSN {
		t.Fatalf("expected ErrInvalidDSN when DATABASE_DSN is empty, got %v", err)
	}
}

func TestInitServerFlags_DSNFlagValid_OK(t *testing.T) {
	clearFlags()
	defer withEnv("DATABASE_DSN", "")()

	os.Args = []string{"cmd", "-d", "postgres://user:pass@localhost:5432/metrics"}

	serverFlags, err := InitServerFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDSN := "postgres://user:pass@localhost:5432/metrics"
	if serverFlags.PostgresDSN != expectedDSN {
		t.Errorf("expected DSN=%s, got %s", expectedDSN, serverFlags.PostgresDSN)
	}
}

func TestInitServerFlags_DSNEnvValid_OK(t *testing.T) {
	clearFlags()
	restore := withEnv("DATABASE_DSN", "postgres://user:pass@db:5432/metrics")
	defer restore()

	os.Args = []string{"cmd"}

	serverFlags, err := InitServerFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDSN := "postgres://user:pass@db:5432/metrics"
	if serverFlags.PostgresDSN != expectedDSN {
		t.Errorf("expected DSN=%s, got %s", expectedDSN, serverFlags.PostgresDSN)
	}
}

func TestInitServerFlags_DSNNotProvided_OK(t *testing.T) {
	clearFlags()
	defer withEnv("DATABASE_DSN", "")()

	os.Args = []string{"cmd"} // без -d и без env

	serverFlags, err := InitServerFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if serverFlags.PostgresDSN != "" {
		t.Errorf("expected empty DSN when not provided, got %s", serverFlags.PostgresDSN)
	}
}
