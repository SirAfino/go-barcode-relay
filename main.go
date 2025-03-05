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

package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sirafino/go-barcode-relay/reader"
	"sirafino/go-barcode-relay/sender"

	"github.com/holoplot/go-evdev"
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
	ID      string                `yaml:"id"`
	Devices []DeviceConfiguration `yaml:"devices"`
	Target  TargetConfiguration   `yaml:"target"`
}

func main() {
	// Whole app configuration
	var config Configuration

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Println("Barcode Relay (Go)")

	// TODO: make this seriously
	if len(os.Args) > 1 && os.Args[1] == "--list" {
		devices, error := reader.ListAvailableDevices()
		if error != nil {
			fmt.Println("Error while reading available devices")
			return
		}

		for i, device := range devices {
			name, error := device.Name()
			if error != nil {
				name = ""
			}

			path := device.Path()

			ids, error := device.InputID()
			if error != nil {
				// TODO:
				continue
			}

			fmt.Printf("%d - %s - %s - VID %04d - PID %04d\n", i, name, path, ids.Vendor, ids.Product)

			evtypes := device.CapableTypes()
			for _, evtype := range evtypes {
				fmt.Println(evdev.TypeName(evtype))
			}
		}

		return
	}

	// Load configuration from a yaml file
	yamlFile, err := os.ReadFile("config/config.yml")
	if err != nil {
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		panic(err)
	}

	// Keep a list of readers, one for each device to be read
	readers := make([]*reader.DeviceReader, len(config.Devices))

	// Instantiate each device reader based on the configuration
	for idx, readerConfig := range config.Devices {
		readers[idx] = nil

		regex, err := regexp.Compile(readerConfig.FullScanRegex)
		if err != nil {
			// TODO
			fmt.Printf("Error on regex")
			continue
		}

		deviceReader := reader.DeviceReader{
			DeviceID: readerConfig.ID,
			VID:      readerConfig.VID,
			PID:      readerConfig.PID,
			Regex:    regex,
		}

		readers[idx] = &deviceReader
	}

	// Create sender
	var s sender.Sender

	switch config.Target.Type {
	case "redis":
		s = &sender.RedisStreamSender{
			Host:     config.Target.Host,
			Port:     config.Target.Port,
			Username: config.Target.Username,
			Password: config.Target.Password,
			Stream:   config.Target.Stream,
		}
	case "dummy":
		s = &sender.DummySender{}
	default:
		s = &sender.DummySender{}
	}

	// Create the scans channel
	scans := make(chan reader.Scan)

	// Start all readers
	for _, reader := range readers {
		go reader.Run(scans, 1000)
	}

	// Start sender
	s.Run(ctx, scans, config.ID)
}
