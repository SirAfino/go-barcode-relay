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

const (
	interceptionDevice = `\\.\interception`

	interceptionIoctlSetFilter = 0x222010
	interceptionIoctlGetFilter = 0x222020
	interceptionIoctlSetEvent  = 0x222040
	interceptionIoctlReceive   = 0x222100
	interceptionIoctlGetHWID   = 0x222200

	keystrokeByteSize = 12

	MaxDevices   = 20 // The Interception driver supports up to 20 devices
	MaxKeyboards = 10

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
