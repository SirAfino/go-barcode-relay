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

package reader

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/holoplot/go-evdev"
)

func ListAvailableDevices() ([]*evdev.InputDevice, error) {
	var devices []*evdev.InputDevice = make([]*evdev.InputDevice, 0)

	paths, error := evdev.ListDevicePaths()
	if error != nil {
		return devices, error
	}

	for _, path := range paths {
		device, error := evdev.Open(path.Path)
		if error != nil {
			// TODO: maybe give some warning here?
			continue
		}

		evtypes := device.CapableTypes()
		if !slices.Contains(evtypes, evdev.EV_KEY) {
			// Skip device if it cannot send "EV_KEY" events
			continue
		}

		devices = append(devices, device)
	}

	return devices, nil
}

func FindDeviceByIDs(vid uint16, pid uint16) (*evdev.InputDevice, error) {
	paths, error := evdev.ListDevicePaths()
	if error != nil {
		return nil, error
	}

	for _, path := range paths {
		device, error := evdev.Open(path.Path)
		if error != nil {
			// TODO: maybe give some warning here?
			continue
		}

		ids, error := device.InputID()
		if error != nil {
			// TODO:
			continue
		}

		if ids.Vendor == vid && ids.Product == pid {
			return device, nil
		}
	}

	return nil, errors.New("notfound")
}

type DeviceReader struct {
	VID         uint16
	PID         uint16
	evdevDevice *evdev.InputDevice
	grabbed     bool
	buffer      string
}

func (deviceReader *DeviceReader) Reset() {
	deviceReader.evdevDevice = nil
	deviceReader.grabbed = false
	deviceReader.buffer = ""
}

func (deviceReader *DeviceReader) Run() {
	regex, err := regexp.Compile("^.*?\n$")
	if err != nil {
		// TODO
		fmt.Printf("Error on regex")
		return
	}

	for {
		if deviceReader.evdevDevice == nil {
			evdevDevice, error := FindDeviceByIDs(deviceReader.VID, deviceReader.PID)
			if error != nil {
				fmt.Printf("Device not found for: VID %04d, PID %04d\n", deviceReader.VID, deviceReader.PID)
				fmt.Printf("Trying again in %d seconds\n", 5)
				time.Sleep(5 * time.Second)
				continue
			}

			deviceReader.Reset()
			deviceReader.evdevDevice = evdevDevice
			fmt.Printf("Device found\n")
		}

		if !deviceReader.grabbed {
			error := deviceReader.evdevDevice.Grab()
			if error != nil {
				deviceReader.Reset()

				fmt.Printf("Error while grabbing device\n")
				fmt.Printf("Trying again in %d seconds\n", 5)
				time.Sleep(5 * time.Second)
				continue
			}
		}

		for {
			event, error := deviceReader.evdevDevice.ReadOne()
			if error != nil {
				// deviceReader.Reset() ???
				break
			}

			if event.Type != evdev.EV_KEY {
				// Don't care about other events
				continue
			}

			if event.Value != 1 {
				// Not a keydown event
				continue
			}

			code := event.Code
			deviceReader.buffer += CharMap[(uint16)(code)]

			if regex.Match([]byte(deviceReader.buffer)) {
				fmt.Printf("Read message: %s\n", deviceReader.buffer)
				deviceReader.buffer = ""
			}
		}
	}
}
