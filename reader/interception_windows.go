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
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

const (
	interceptionDevice = `\\.\interception`

	interceptionIoctlSetFilter = 0x222010
	interceptionIoctlGetFilter = 0x222020
	interceptionIoctlSetEvent  = 0x222040
	interceptionIoctlReceive   = 0x222100
	interceptionIoctlGetHWID   = 0x222200

	maxDevices   = 20 // The Interception driver supports up to 20 devices
	maxKeyboards = 10

	INTERCEPTION_KEY_DOWN             = 0x00
	INTERCEPTION_KEY_UP               = 0x01
	INTERCEPTION_KEY_E0               = 0x02
	INTERCEPTION_KEY_E1               = 0x04
	INTERCEPTION_KEY_TERMSRV_SET_LED  = 0x08
	INTERCEPTION_KEY_TERMSRV_SHADOW   = 0x10
	INTERCEPTION_KEY_TERMSRV_VKPACKET = 0x20

	INTERCEPTION_FILTER_KEY_NONE             = 0x0000
	INTERCEPTION_FILTER_KEY_ALL              = 0xFFFF
	INTERCEPTION_FILTER_KEY_DOWN             = INTERCEPTION_KEY_UP
	INTERCEPTION_FILTER_KEY_UP               = INTERCEPTION_KEY_UP << 1
	INTERCEPTION_FILTER_KEY_E0               = INTERCEPTION_KEY_E0 << 1
	INTERCEPTION_FILTER_KEY_E1               = INTERCEPTION_KEY_E1 << 1
	INTERCEPTION_FILTER_KEY_TERMSRV_SET_LED  = INTERCEPTION_KEY_TERMSRV_SET_LED << 1
	INTERCEPTION_FILTER_KEY_TERMSRV_SHADOW   = INTERCEPTION_KEY_TERMSRV_SHADOW << 1
	INTERCEPTION_FILTER_KEY_TERMSRV_VKPACKET = INTERCEPTION_KEY_TERMSRV_VKPACKET << 1
)

func isKeyboard(device int) bool {
	return device+1 > 0 && device+1 <= maxKeyboards
}

func DecodeUtf16(b []byte, order binary.ByteOrder) (string, error) {
	ints := make([]uint16, len(b)/2)
	if err := binary.Read(bytes.NewReader(b), order, &ints); err != nil {
		return "", err
	}
	return string(utf16.Decode(ints)), nil
}

type InterceptionDevice struct {
	index  int
	name   string
	handle windows.Handle
	event  windows.Handle
}

func NewDevice(index int) (*InterceptionDevice, error) {
	name := fmt.Sprintf("%s%02d", interceptionDevice, index)
	namePtr, err := windows.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}

	handle, err := windows.CreateFile(
		namePtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		0,
		0,
	)
	if err != nil {
		return nil, err
	}

	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return nil, err
	}

	device := InterceptionDevice{
		index:  index,
		name:   name,
		handle: handle,
		event:  event,
	}

	return &device, nil
}

func (device *InterceptionDevice) Init() error {
	var buffer []byte = make([]byte, 128)
	var bytesReturned uint32

	binary.LittleEndian.PutUint64(buffer, uint64(device.event))

	err := windows.DeviceIoControl(
		device.handle,
		uint32(interceptionIoctlSetEvent),
		&buffer[0],
		uint32(len(buffer)),
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		return fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	return nil
}

func (device *InterceptionDevice) GetHWID() (string, error) {
	var buffer [1024]byte

	var bytesReturned uint32

	err := windows.DeviceIoControl(
		device.handle,
		uint32(interceptionIoctlGetHWID),
		nil,
		0,
		&buffer[0],
		uint32(len(buffer)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	hwidStr, err := DecodeUtf16(buffer[:bytesReturned], binary.LittleEndian)
	if err != nil {
		return "", err
	}

	return hwidStr, nil
}

func (device *InterceptionDevice) GetFilter() (*uint16, error) {
	var buffer [1024]byte
	var bytesReturned uint32

	err := windows.DeviceIoControl(
		device.handle,
		uint32(interceptionIoctlGetFilter),
		nil,
		0,
		&buffer[0],
		uint32(len(buffer)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	var filter uint16 = 0
	if bytesReturned >= 2 {
		filter = binary.LittleEndian.Uint16(buffer[:2])
	}

	return &filter, nil
}

func (device *InterceptionDevice) SetFilter(filter uint16) error {
	var buffer []byte = make([]byte, 2)
	var bytesReturned uint32

	binary.BigEndian.PutUint16(buffer, filter)

	err := windows.DeviceIoControl(
		device.handle,
		uint32(interceptionIoctlSetFilter),
		&buffer[0],
		uint32(len(buffer)),
		nil,
		0,
		&bytesReturned,
		nil,
	)
	if err != nil {
		return fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	return nil
}

func (device *InterceptionDevice) Wait(timeout uint32) (uint32, error) {
	event, err := windows.WaitForSingleObject(device.event, timeout)
	return event, err
}

func (device *InterceptionDevice) Receive() ([]byte, error) {
	var buffer [1024]byte

	var bytesReturned uint32

	err := windows.DeviceIoControl(
		device.handle,
		uint32(interceptionIoctlReceive),
		nil,
		0,
		&buffer[0],
		uint32(len(buffer)),
		&bytesReturned,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	return buffer[:bytesReturned], nil
}

func (device *InterceptionDevice) Close() error {
	err := windows.CloseHandle(device.handle)
	if err != nil {
		return err
	}

	return windows.CloseHandle(device.event)
}

type Interception struct {
	devices []*InterceptionDevice
}

var interception Interception = Interception{
	devices: make([]*InterceptionDevice, maxDevices),
}

func GetInterception() *Interception {
	return &interception
}

func (interception *Interception) Init() {
	for i := range maxDevices {
		if interception.devices[i] != nil {
			interception.devices[i].Close()
		}

		device, err := NewDevice(i)
		if err != nil {
			continue
		}

		err = device.Init()
		if err != nil {
			fmt.Printf("Error while initializing device: %s\n", err)
		}

		interception.devices[i] = device
	}
}

func (interception *Interception) FindDeviceByIDs(vid uint16, pid uint16) (*InterceptionDevice, error) {
	vidStr := fmt.Sprintf("VID_%04X", vid)
	pidStr := fmt.Sprintf("PID_%04X", pid)

	for _, device := range interception.devices {
		hwid, err := device.GetHWID()
		if err != nil {
			// TODO
			continue
		}

		if strings.Contains(hwid, vidStr) && strings.Contains(hwid, pidStr) {
			return device, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

// ListDevices retrieves a list of active input devices (keyboards/mice)
func ListDevices() ([]string, error) {
	var results []string = make([]string, 0)

	for i := range maxDevices {
		if !isKeyboard(i) {
			continue
		}

		device, err := NewDevice(i)
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
