package test

import (
	"fmt"
	"math/rand"
	"sirafino/go-barcode-relay/configuration"
	"sirafino/go-barcode-relay/logging"
	"sirafino/go-barcode-relay/reader"
	"sirafino/go-barcode-relay/sender"
	"strings"
	"sync"
	"testing"
	"time"
)

const letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytes(n int) string {
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}

func simulateReader(scans chan reader.Scan, deviceID string, wg *sync.WaitGroup) {
	defer wg.Done()

	const scansCount = 100
	const scanSize = 8
	const minWaitTime = 10
	const maxWaitTime = 50

	for range scansCount {
		scan := RandStringBytes(scanSize)
		scans <- reader.Scan{
			DeviceID:  deviceID,
			Content:   scan,
			Timestamp: time.Now().Unix(),
		}

		time.Sleep(time.Duration(rand.Int()%(maxWaitTime-minWaitTime)+minWaitTime) * time.Millisecond)
	}
}

func TestSenderStressTest(t *testing.T) {
	// Setup Redis connection
	logger := logging.GetLogger("APP")

	// Whole app configuration
	var config *configuration.Configuration
	var err error

	config, err = configuration.LoadConfiguration("../config/config.yml")
	if err != nil {
		logger.Error("Error while loading configuration file")
		panic(err)
	}
	logger.Info("Configuration file loaded (%d device/s, %d target/s)", len(config.Devices), 1)

	// Create sender
	var s sender.Sender

	switch config.Target.Type {
	case "redis":
		s = &sender.RedisStreamSender{
			Host:     config.Target.Host,
			Port:     config.Target.Port,
			Username: config.Target.Username,
			Password: config.Target.Password,
			Stream:   config.Target.Stream,
		}
	case "dummy":
		s = &sender.DummySender{}
	default:
		s = &sender.DummySender{}
	}

	// Create the scans channel
	scans := make(chan reader.Scan)

	// Create waitgroups for readers and senders
	var readersWaitGroup sync.WaitGroup
	var sendersWaitGroup sync.WaitGroup

	// Start sender
	sendersWaitGroup.Add(1)
	go s.Run(scans, config.ID, &sendersWaitGroup)
	logger.Info("Sender/s started")

	for i := range 10 {
		readersWaitGroup.Add(1)
		go simulateReader(scans, fmt.Sprintf("reader%02d", i), &readersWaitGroup)
	}

	readersWaitGroup.Wait()

	close(scans)

	sendersWaitGroup.Wait()
}
