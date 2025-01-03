package nats

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

var Conn *nats.Conn

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
	Conn = c
}

func Publish(subject string, data []byte) {
	err := Conn.Publish(subject, data)
	if err != nil {
		log.WithError(err).Error("Failed to publish message")
	}
}
