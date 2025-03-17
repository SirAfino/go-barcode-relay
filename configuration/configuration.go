//
// This file is part of the GoBarcodeRelay distribution (https://github.com/SirAfino/go-barcode-relay).
// Copyright (c) 2025 Gabriele Serafino.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
//

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
