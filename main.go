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
	"fmt"
	"os"
	"regexp"
	"sirafino/go-barcode-relay/reader"

	"github.com/holoplot/go-evdev"
)

func main() {
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

	// TODO: this is for testing only
	var vid uint16 = 1452
	var pid uint16 = 591

	regex, err := regexp.Compile("^.*?\n$")
	if err != nil {
		// TODO
		fmt.Printf("Error on regex")
		return
	}

	deviceReader := reader.DeviceReader{
		VID:   vid,
		PID:   pid,
		Regex: regex,
	}

	scans := make(chan string)

	go deviceReader.Run(scans, 1000)

	for scan := range scans {
		fmt.Printf("Read message: %s\n", scan)
	}
}
