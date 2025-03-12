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
	"context"
	"errors"
	"fmt"
	"regexp"
	"sirafino/go-barcode-relay/logging"
	"slices"
	"strings"
	"sync"
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
	DeviceID    string
	VID         uint16
	PID         uint16
	Regex       *regexp.Regexp
	evdevDevice *evdev.InputDevice
	grabbed     bool
	buffer      string
	logger      *logging.Logger
}

func (deviceReader *DeviceReader) Reset() {
	deviceReader.evdevDevice = nil
	deviceReader.grabbed = false
	deviceReader.buffer = ""
}

func (deviceReader *DeviceReader) readCharacter() (*string, error) {
	event, err := deviceReader.evdevDevice.ReadOne()

	if err != nil {
		deviceReader.logger.Info("Device disconnected\n")
		deviceReader.Reset()
		return nil, fmt.Errorf("disconnected")
	}

	if event.Type != evdev.EV_KEY {
		// Don't care about other events
		return nil, nil
	}

	if event.Value != 1 {
		// Not a keydown event
		return nil, nil
	}

	code := event.Code

	character := CharMap[(uint16)(code)]

	return &character, nil
}

func (deviceReader *DeviceReader) readCharacters(characters chan string, polling_ms int16) {
	for {
		if deviceReader.evdevDevice == nil {
			evdevDevice, error := FindDeviceByIDs(deviceReader.VID, deviceReader.PID)
			if error != nil {
				time.Sleep(time.Duration(polling_ms) * time.Millisecond)
				continue
			}

			deviceReader.logger.Info("Device connected\n")
			deviceReader.Reset()
			deviceReader.evdevDevice = evdevDevice
		}

		if !deviceReader.grabbed {
			error := deviceReader.evdevDevice.Grab()
			if error != nil {
				deviceReader.logger.Error("Error while grabbing device, trying again in %d ms\n", polling_ms)
				deviceReader.Reset()
				time.Sleep(time.Duration(polling_ms) * time.Millisecond)
				continue
			}

			deviceReader.grabbed = true
		}

		for {
			character, err := deviceReader.readCharacter()
			if err != nil {
				break
			}

			if character != nil {
				characters <- *character
			}
		}
	}
}

func (deviceReader *DeviceReader) Run(
	ctx context.Context,
	scans chan Scan,
	polling_ms int16,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	if deviceReader.logger == nil {
		deviceReader.logger = logging.GetLogger("READER:" + deviceReader.DeviceID)
	}

	characters := make(chan string, 1)

	go deviceReader.readCharacters(characters, polling_ms)

	for {
		select {
		case <-ctx.Done():
			deviceReader.logger.Info("Stopping device reader: %s", deviceReader.DeviceID)
			return
		case character := <-characters:
			deviceReader.buffer += character

			if deviceReader.Regex.Match([]byte(deviceReader.buffer)) {
				scan := Scan{
					DeviceID:  deviceReader.DeviceID,
					Content:   deviceReader.buffer,
					Timestamp: time.Now().Unix(),
				}

				deviceReader.logger.Info("Read scan (%s)", strings.ReplaceAll(scan.Content, "\n", ""))

				scans <- scan
				deviceReader.buffer = ""
			}
		}
	}
}
