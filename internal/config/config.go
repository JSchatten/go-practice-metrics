package config

import (
	"encoding/json"
	"os"
)

// ServerConfig содержит конфигурацию сервера метрик.
type ServerConfig struct {
	Address        string `json:"address"`
	Restore        bool   `json:"restore"`
	StoreInterval  string `json:"store_interval"`
	StoreFile      string `json:"store_file"`
	DatabaseDSN    string `json:"database_dsn"`
	CryptoKey      string `json:"crypto_key"`
	TrustedSubnet  string `json:"trusted_subnet"`
	GRPCServerPort uint   `json:"grpc_server_port"`
}

// AgentConfig содержит конфигурацию агента метрик.
type AgentConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
	GRPCServerAddr string `json:"grpc_serrver_addr"`
}

// LoadServerConfig загружает конфигурацию сервера из JSON-файла.
func LoadServerConfig(path string) (*ServerConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config ServerConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadAgentConfig загружает конфигурацию агента из JSON-файла.
func LoadAgentConfig(path string) (*AgentConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config AgentConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
