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

//go:build windows

package reader

import (
	"context"
	"fmt"
	"regexp"
	"sirafino/go-barcode-relay/interception"
	"sirafino/go-barcode-relay/logging"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

// ListDevices retrieves a list of active input devices (keyboards/mice)
func ListDevices() ([]string, error) {
	var results []string = make([]string, 0)

	for i := range interception.MaxDevices {
		if !interception.IsKeyboard(i) {
			continue
		}

		device, err := interception.NewDevice(i)
		if err != nil {
			continue
		}

		hwid, err := device.GetHWID()
		if err != nil {
			continue
		}

		results = append(results, hwid)
	}

	return results, nil
}

type DeviceReader struct {
	DeviceID string
	VID      uint16
	PID      uint16
	Regex    *regexp.Regexp
	device   *interception.Device
	buffer   string
	logger   *logging.Logger
}

func (deviceReader *DeviceReader) findDevice() bool {
	device, err := interception.FindDeviceByIDs(deviceReader.VID, deviceReader.PID)
	if err != nil {
		return false
	}

	// Set filter for all event types, even if we only need keydowns, this prevents
	// the other event types to interact with the rest of os and apps
	err = device.SetFilter(interception.INTERCEPTION_FILTER_KEY_ALL)
	if err != nil {
		deviceReader.logger.Error("Unable to set filter for device: %s", err)
		return false
	}

	deviceReader.device = device
	return true
}

func (deviceReader *DeviceReader) readCharacters(characters chan string, polling_ms int16) {
	for {
		if deviceReader.device == nil {
			// Try to get the device
			found := deviceReader.findDevice()
			if !found {
				// If not found, just wait some time and try again
				time.Sleep(5000 * time.Millisecond)
				continue
			}

			deviceReader.logger.Info("Device connected\n")
		}

		// Wait for an event from the device. Here we cannot wait indefenetely since
		// if the device is unplugged, no event will be triggered and we would be stuck
		// at the "wait" call.
		res, _ := deviceReader.device.Wait(uint32(polling_ms))

		// Mostly two cases here:
		//  1 - the "wait" call returns because an event has been signaled, probably a keystroke
		//  2 - the "wait" call has timed out

		// If wait has timed out, try to read HWID from the device, if the device has
		// been disconnected, this will result in an error.
		if res != uint32(windows.WAIT_OBJECT_0) {
			_, err := deviceReader.device.GetHWID()
			if err != nil {
				// The device has probably disconnected,
				// lose handle to the device so that next iteration will try to find it again
				deviceReader.logger.Info("Device disconnected\n")
				deviceReader.device.Close()
				deviceReader.device = nil
			}

			// The device has not disconnected, just no event was fired during the timeout,
			// Keep waiting for events on next iteration.
			continue
		}

		// The "wait" call has returned after an event has been received, try to read next
		// keystroke from device
		keystroke, err := deviceReader.device.Receive()
		if err != nil {
			fmt.Printf("Error while receiving from device: %s\n", err)
			continue
		}

		// Skip everything that is not a keydown event
		if keystroke.State != interception.INTERCEPTION_KEY_DOWN {
			continue
		}

		// Convert the keystroke code to a character using the CharMap
		character := CharMap[(uint16)(keystroke.Code)]

		characters <- character
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

	// Start a new goroutine that reads from the device, one character at a time
	go deviceReader.readCharacters(characters, polling_ms)

	for {
		select {
		case <-ctx.Done():
			deviceReader.logger.Info("Stopping device reader: %s", deviceReader.DeviceID)
			return
		case character := <-characters:
			// Append the new character to the buffer
			deviceReader.buffer += character

			// Check if the buffer matches the full_scan_regex
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
