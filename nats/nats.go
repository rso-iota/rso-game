package nats

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

var conn *nats.Conn

func Connect(url string) {
	if url == "" {
		// No NATS server configured, do nothing.
		log.Info("No nats server configured")
		return
	}

	c, err := nats.Connect(url)
	if err != nil {
		log.WithError(err).Error("Failed to connect to nats")
		return
	}

	log.Info("Connected to nats at ", url)
	conn = c
}

func Publish(subject string, data []byte) {
	err := conn.Publish(subject, data)
	if err != nil {
		log.WithError(err).Error("Failed to publish message")
	}
}
