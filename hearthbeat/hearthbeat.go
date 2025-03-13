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

package hearthbeat

import (
	"context"
	"encoding/json"
	"time"
)

var startTime time.Time = time.Now()

type HearthbeatConfiguration struct {
	Type     string `yaml:"type"`
	Interval int    `yaml:"interval"`
}

type Hearthbeat interface {
	// Start the hearthbeat thread. Send hearthbeat and some telemetry at a specific interval
	Run(ctx context.Context, relayID string)
}

func getHearthbeatMessage(relayID string) map[string]any {
	devices := map[string]any{} // TODO
	devicesJson, _ := json.Marshal(devices)

	return map[string]any{
		"relay":   relayID,
		"uptime":  int(time.Since(startTime).Seconds()),
		"devices": devicesJson,
		"ts":      time.Now().Unix(),
	}
}
