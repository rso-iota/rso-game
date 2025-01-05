package nats

import (
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

var conn *nats.Conn
var url string

func Connect(natsUrl string) {
	url = natsUrl
	if natsUrl == "" {
		// No NATS server configured, do nothing.
		log.Info("No nats server configured")
		return
	}

	c, err := nats.Connect(natsUrl)
	if err != nil {
		log.WithError(err).Error("Failed to connect to nats")
		return
	}

	log.Info("Connected to nats at ", natsUrl)
	conn = c
}

func Publish(subject string, data []byte) {
	err := conn.Publish(subject, data)
	if err != nil {
		log.WithError(err).Error("Failed to publish message")
	}
}
