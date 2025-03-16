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
	"encoding/binary"
	"fmt"
	"regexp"
	"sirafino/go-barcode-relay/logging"
	"strings"
	"sync"
	"time"

	"golang.org/x/sys/windows"
)

type DeviceReader struct {
	DeviceID string
	VID      uint16
	PID      uint16
	Regex    *regexp.Regexp
	device   *InterceptionDevice
	buffer   string
	logger   *logging.Logger
}

func (deviceReader *DeviceReader) readCharacters(characters chan string, polling_ms int16) {
	for {
		if deviceReader.device == nil {
			interception := GetInterception()
			interception.Init()

			device, err := interception.FindDeviceByIDs(deviceReader.VID, deviceReader.PID)
			if err != nil {
				deviceReader.logger.Error("Error while getting device: %s", err)
				time.Sleep(5000 * time.Millisecond)
				continue // TODO
			}

			err = device.SetFilter(INTERCEPTION_FILTER_KEY_ALL)
			if err != nil {
				deviceReader.logger.Error("Unable to set filter for device: %s", err)
				time.Sleep(5000 * time.Millisecond)
				continue
			}

			deviceReader.device = device
		}

		res, _ := deviceReader.device.Wait(uint32(2000))
		fmt.Printf("%d - ", res)

		if res != uint32(windows.WAIT_OBJECT_0) {
			_, err := deviceReader.device.GetHWID()
			if err != nil {
				// The device has probably disconnected
				deviceReader.device.Close()
				deviceReader.device = nil
			}

			continue
		}

		event, err := deviceReader.device.Receive()
		if err != nil {
			fmt.Printf("Error while receiving from device: %s\n", err)
			continue
		}

		v1 := binary.LittleEndian.Uint16(event[:2])
		code := binary.LittleEndian.Uint16(event[2:4])
		state := binary.LittleEndian.Uint16(event[4:6])
		v4 := binary.LittleEndian.Uint16(event[6:8])
		information := binary.LittleEndian.Uint32(event[8:])

		// Skip everything that is not a keydown event
		if state != INTERCEPTION_KEY_DOWN {
			continue
		}

		character := CharMap[(uint16)(code)]

		characters <- character

		fmt.Printf("%d %d %d %d %d\n", v1, code, state, v4, information)
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
