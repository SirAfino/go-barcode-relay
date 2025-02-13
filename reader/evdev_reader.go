package reader

import (
	"errors"
	"fmt"
	"slices"

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

func ReadFromDevice(device *evdev.InputDevice) {
	var error error

	error = device.Grab()
	if error != nil {
		fmt.Printf("Error while grabbing device")
		return
	}

	for {
		event, error := device.ReadOne()
		if error != nil {
			continue
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

		fmt.Printf("%s\n", CharMap[(uint16)(code)])
	}
}
