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
	ctx context.Context,
	scans chan reader.Scan,
	relayID string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

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

	for {
		select {
		case <-ctx.Done():
			if len(scans) == 0 {
				sender.logger.Info("Stopping sender\n")
				return
			}
		case scan := <-scans:
			sender.logger.Info("Sent message: (%s)\n", strings.ReplaceAll(scan.Content, "\n", ""))
			client.XAdd(ctx, &redis.XAddArgs{
				Stream: sender.Stream,
				Values: map[string]any{
					"relay":  relayID,
					"device": scan.DeviceID,
					"code":   scan.Content,
					"ts":     scan.Timestamp,
				},
			})
		}
	}
}
