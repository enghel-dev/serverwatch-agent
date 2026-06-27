package config

import (
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// AgentConfig representa la estructura de la configuración del agente
type AgentConfig struct {
	BackendHost string `yaml:"backend_host"`
	DisplayName string `yaml:"display_name"`
	AgentUUID   string `yaml:"agent_uuid"`
}

// GetConfigPath devuelve la ruta del archivo de configuración dependiendo del sistema operativo
func GetConfigPath() string{
	if runtime.GOOS == "windows" {
		return "C:\\ProgramData\\ServerWatch\\config.yaml"
	}
	return "/etc/serverwatch/config.yaml"
}

// ConfigExists verifica si el archivo de configuración existe en la ruta especificada
func ConfigExists() bool {
	path := GetConfigPath()
	_, err := os.Stat(path)
	return err == nil
}

// LoadConfig carga la configuración desde el archivo YAML y la devuelve como una estructura AgentConfig
func LoadConfig() (*AgentConfig, error) {
	path := GetConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AgentConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveConfig guarda la configuración del agente en un archivo YAML en la ruta especificada
func SaveConfig(cfg *AgentConfig) error {
	path := GetConfigPath()
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}