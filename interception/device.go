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

package interception

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

func decodeUtf16(b []byte, order binary.ByteOrder) (string, error) {
	ints := make([]uint16, len(b)/2)
	if err := binary.Read(bytes.NewReader(b), order, &ints); err != nil {
		return "", err
	}
	return string(utf16.Decode(ints)), nil
}

type Device struct {
	index  int
	name   string
	handle windows.Handle
	event  windows.Handle

	inputBuffer   [1024]byte
	outputBuffer  [1024]byte
	bytesReturned uint32
}

func NewDevice(index int) (*Device, error) {
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

	device := Device{
		index:  index,
		name:   name,
		handle: handle,
		event:  event,
	}

	return &device, nil
}

func (device *Device) Init() error {
	binary.LittleEndian.PutUint64(device.inputBuffer[:], uint64(device.event))

	err := device.deviceIoControl(interceptionIoctlSetEvent)
	if err != nil {
		return fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	return nil
}

func (device *Device) GetHWID() (string, error) {
	err := device.deviceIoControl(interceptionIoctlGetHWID)
	if err != nil {
		return "", fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	hwidStr, err := decodeUtf16(device.outputBuffer[:device.bytesReturned], binary.LittleEndian)
	if err != nil {
		return "", err
	}

	return hwidStr, nil
}

func (device *Device) GetFilter() (*uint16, error) {
	err := device.deviceIoControl(interceptionIoctlGetFilter)
	if err != nil {
		return nil, fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	var filter uint16 = 0
	if device.bytesReturned >= 2 {
		filter = binary.LittleEndian.Uint16(device.outputBuffer[:2])
	}

	return &filter, nil
}

func (device *Device) SetFilter(filter uint16) error {
	binary.BigEndian.PutUint16(device.inputBuffer[:], filter)

	err := device.deviceIoControl(interceptionIoctlSetFilter)
	if err != nil {
		return fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	return nil
}

func (device *Device) Wait(timeout uint32) (uint32, error) {
	event, err := windows.WaitForSingleObject(device.event, timeout)
	return event, err
}

func (device *Device) Receive() ([]*KeyStroke, error) {
	err := device.deviceIoControl(interceptionIoctlReceive)
	if err != nil {
		return nil, fmt.Errorf("call to DeviceIoControl failed: %v", err)
	}

	if device.bytesReturned < keystrokeByteSize {
		return nil, fmt.Errorf("invalid read, not enough data")
	}

	strokes := make([]*KeyStroke, int(device.bytesReturned/keystrokeByteSize))
	for i := range int(device.bytesReturned / keystrokeByteSize) {
		buffer := device.outputBuffer[i*keystrokeByteSize : i*keystrokeByteSize+12]
		// v1 := binary.LittleEndian.Uint16(buffer[:2])
		code := binary.LittleEndian.Uint16(buffer[2:4])
		state := binary.LittleEndian.Uint16(buffer[4:6])
		// v4 := binary.LittleEndian.Uint16(buffer[6:8])
		information := binary.LittleEndian.Uint32(buffer[8:12])

		strokes[i] = &KeyStroke{
			code,
			state,
			information,
		}
	}

	return strokes, nil
}

func (device *Device) Close() error {
	err := windows.CloseHandle(device.handle)
	if err != nil {
		return err
	}

	return windows.CloseHandle(device.event)
}

func (device *Device) deviceIoControl(ioControlCode uint32) error {
	return windows.DeviceIoControl(
		device.handle,
		ioControlCode,
		&device.inputBuffer[0],
		uint32(len(device.inputBuffer)),
		&device.outputBuffer[0],
		uint32(len(device.outputBuffer)),
		&device.bytesReturned,
		nil,
	)
}
