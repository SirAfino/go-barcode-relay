package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/holoplot/go-evdev"
)

var charMap = map[uint16]string{
	2:  "1",
	3:  "2",
	4:  "3",
	5:  "4",
	6:  "5",
	7:  "6",
	8:  "7",
	9:  "8",
	10: "9",
	11: "0",
	16: "Q",
	17: "W",
	18: "E",
	19: "R",
	20: "T",
	21: "Y",
	22: "U",
	23: "I",
	24: "O",
	25: "P",
	28: "\n", // Enter
	30: "A",
	31: "S",
	32: "D",
	33: "F",
	34: "G",
	35: "H",
	36: "J",
	37: "K",
	38: "L",
	42: "", // Shift
	44: "Z",
	45: "X",
	46: "C",
	47: "V",
	48: "B",
	49: "N",
	50: "M",
}

func listAvailableDevices() ([]*evdev.InputDevice, error) {
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

func findDeviceByIDs(vid uint16, pid uint16) (*evdev.InputDevice, error) {
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

func main() {
	fmt.Println("Barcode Relay (Go)")

	// TODO: make this seriously
	if len(os.Args) > 1 && os.Args[1] == "--list" {
		devices, error := listAvailableDevices()
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

	device, error := findDeviceByIDs(vid, pid)
	if error != nil {
		fmt.Printf("Device not found for: VID %04d, PID %04d\n", pid, vid)
		return
	}

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

		fmt.Printf("%s\n", charMap[(uint16)(code)])
	}

}
