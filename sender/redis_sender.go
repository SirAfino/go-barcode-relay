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
	"sirafino/go-barcode-relay/logging"
	"sirafino/go-barcode-relay/reader"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStreamSender struct {
	Host     string
	Port     int16
	Username string
	Password string
	Stream   string
	logger   *logging.Logger
}

func (sender *RedisStreamSender) Run(
	scans chan reader.Scan,
	relayID string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	ctx := context.Background()

	if sender.logger == nil {
		sender.logger = logging.GetLogger("SENDER")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", sender.Host, sender.Port),
		Username: sender.Username,
		Password: sender.Password,
		DB:       0, // Use default DB
		Protocol: 2, // Connection protocol
	})

	var scan *reader.Scan

	for {
		// If current scan is nil, read the next scan to send from the channel
		if scan == nil {
			s, ok := <-scans
			if !ok {
				sender.logger.Info("Stopping sender\n")
				return
			}

			scan = &s
		}

		cmd := client.XAdd(ctx, &redis.XAddArgs{
			Stream: sender.Stream,
			Values: map[string]any{
				"relay":  relayID,
				"device": scan.DeviceID,
				"code":   scan.Content,
				"ts":     scan.Timestamp,
			},
		})

		err := cmd.Err()

		if err != nil {
			// DO NOT clear the scan, so that the next iteration will retry to send this scan

			// Wait some time before retrying
			time.Sleep(5000 * time.Millisecond) // TODO: this could be configurable

			sender.logger.Error("Failed to send message: (%s, %s)\n", strings.ReplaceAll(scan.Content, "\n", ""), err)
		} else {
			sender.logger.Info("Sent message: (%s)\n", strings.ReplaceAll(scan.Content, "\n", ""))

			// Clear scan, so that the next iteration will fetch a new scan from the channel
			scan = nil
		}
	}
}
