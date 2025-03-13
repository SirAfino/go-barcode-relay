package configuration

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DeviceConfiguration struct {
	ID            string `yaml:"id"`
	VID           uint16 `yaml:"vid"`
	PID           uint16 `yaml:"pid"`
	FullScanRegex string `yaml:"full_scan_regex"`
}

type TargetConfiguration struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int16  `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Stream   string `yaml:"stream"`
}

type Configuration struct {
	ID         string                `yaml:"id"`
	Devices    []DeviceConfiguration `yaml:"devices"`
	Target     TargetConfiguration   `yaml:"target"`
	Hearthbeat map[string]any        `yaml:"hearthbeat"`
}

func LoadConfiguration(path string) (*Configuration, error) {
	var config Configuration

	// Load configuration from a yaml file
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
