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
