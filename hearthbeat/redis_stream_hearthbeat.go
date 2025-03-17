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
	"fmt"
	"sirafino/go-barcode-relay/logging"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStreamHearthbeatConfiguration struct {
	Type     string `yaml:"type"`
	Interval int    `yaml:"interval"`

	Host     string `yaml:"host"`
	Port     int16  `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Stream   string `yaml:"stream"`
}

type RedisStreamHearthbeat struct {
	Host     string
	Port     int16
	Username string
	Password string
	Stream   string
	Interval int
	logger   *logging.Logger
}

func (hb *RedisStreamHearthbeat) Run(
	ctx context.Context,
	relayID string,
) {
	if hb.logger == nil {
		hb.logger = logging.GetLogger("HB")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", hb.Host, hb.Port),
		Username: hb.Username,
		Password: hb.Password,
		DB:       0, // Use default DB
		Protocol: 2, // Connection protocol
	})

	ticker := time.NewTicker(time.Duration(hb.Interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			hb.logger.Info("Stopping hearthbeat")
			return
		case <-ticker.C:
			cmd := client.XAdd(ctx, &redis.XAddArgs{
				Stream: hb.Stream,
				Values: getHearthbeatMessage(relayID),
			})

			err := cmd.Err()

			if err != nil {
				hb.logger.Error("Failed to send hearthbeat message: (%s)\n", err)
			}
		}
	}
}
