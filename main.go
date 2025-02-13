package main

import (
	"fmt"
	"os"
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

	device, error := reader.FindDeviceByIDs(vid, pid)
	if error != nil {
		fmt.Printf("Device not found for: VID %04d, PID %04d\n", pid, vid)
		return
	}

	go reader.ReadFromDevice(device)

	for {
	}
}
