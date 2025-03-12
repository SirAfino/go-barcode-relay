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
	"os/signal"
	"regexp"
	"sirafino/go-barcode-relay/configuration"
	"sirafino/go-barcode-relay/logging"
	"sirafino/go-barcode-relay/reader"
	"sirafino/go-barcode-relay/sender"
	"sync"

	"github.com/holoplot/go-evdev"
)

const VERSION string = "1.0.0"

func main() {
	fmt.Println(
		"BarcodeRelay (Go) Copyright (C) 2025  Gabriele Serafino",
		"\nThis program comes with ABSOLUTELY NO WARRANTY.",
		"\nThis is free software, and you are welcome to redistribute it",
		"\nunder certain conditions.",
	)
	fmt.Println()

	logger := logging.GetLogger("APP")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("BarcodeRelay (Go) - v%s", VERSION)

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

	// Whole app configuration
	var config *configuration.Configuration
	var err error

	config, err = configuration.LoadConfiguration("config/config.yml")
	if err != nil {
		logger.Error("Error while loading configuration file")
		panic(err)
	}
	logger.Info("Configuration file loaded (%d device/s, %d target/s)", len(config.Devices), 1)

	// Keep a list of readers, one for each device to be read
	readers := make([]*reader.DeviceReader, len(config.Devices))

	// Instantiate each device reader based on the configuration
	for idx, readerConfig := range config.Devices {
		readers[idx] = nil

		regex, err := regexp.Compile(readerConfig.FullScanRegex)
		if err != nil {
			logger.Error("Invalid regex for device (%s)", readerConfig.ID)
			panic(err)
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

	// Create waitgroups for readers and senders
	var readersWaitGroup sync.WaitGroup
	var sendersWaitGroup sync.WaitGroup

	// Start all readers
	for _, reader := range readers {
		readersWaitGroup.Add(1)
		go reader.Run(ctx, scans, 1000, &readersWaitGroup)
	}
	logger.Info("Reader/s started")

	// Start sender
	sendersWaitGroup.Add(1)
	go s.Run(ctx, scans, config.ID, &sendersWaitGroup)
	logger.Info("Sender/s started")

	// Keep running until a SIGINT is received
	// Setup a channel to receive a signal
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	for range done {
		logger.Info("Received SIGINT, waiting for routines to finish")

		cancel()

		sendersWaitGroup.Wait()
		readersWaitGroup.Wait()

		return
	}
}
