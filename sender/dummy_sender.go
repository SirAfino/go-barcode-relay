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

package sender

import (
	"context"
	"fmt"
	"sirafino/go-barcode-relay/reader"
)

type DummySender struct {
}

func (sender *DummySender) Run(ctx context.Context, scans chan reader.Scan, relayID string) {
	for scan := range scans {
		fmt.Printf("Sent dummy message: %s\n", scan.Content)
	}
}
