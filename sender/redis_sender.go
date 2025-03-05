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

	"github.com/redis/go-redis/v9"
)

type RedisStreamSender struct {
	Host     string
	Port     int16
	Username string
	Password string
	Stream   string
}

func (sender *RedisStreamSender) Run(ctx context.Context, scans chan reader.Scan, relayID string) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", sender.Host, sender.Port),
		Username: sender.Username,
		Password: sender.Password,
		DB:       0, // Use default DB
		Protocol: 2, // Connection protocol
	})

	for scan := range scans {
		fmt.Printf("Sent message: %s\n", scan.Content)
		client.XAdd(ctx, &redis.XAddArgs{
			Stream: sender.Stream,
			Values: map[string]interface{}{
				"relay":  relayID,
				"device": scan.DeviceID,
				"code":   scan.Content,
				"ts":     scan.Timestamp,
			},
		})
	}
}
