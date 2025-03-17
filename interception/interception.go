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
	"fmt"
	"strings"
)

func IsKeyboard(device int) bool {
	return device+1 > 0 && device+1 <= MaxKeyboards
}

func FindDeviceByIDs(vid uint16, pid uint16) (*Device, error) {
	vidStr := fmt.Sprintf("VID_%04X", vid)
	pidStr := fmt.Sprintf("PID_%04X", pid)

	for i := range MaxDevices {
		device, err := NewDevice(i)
		if err != nil {
			continue
		}

		err = device.Init()
		if err != nil {
			fmt.Printf("Error while initializing device: %s\n", err)
		}

		hwid, err := device.GetHWID()
		if err != nil {
			// TODO
			continue
		}

		if strings.Contains(hwid, vidStr) && strings.Contains(hwid, pidStr) {
			return device, nil
		}

		device.Close()
	}

	return nil, fmt.Errorf("not found")
}
