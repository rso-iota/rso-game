package nats

import (
	"rso-game/circuitbreaker"

	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
)

var conn *nats.Conn
var url string

func Connect(natsUrl string) {
	url = natsUrl

	c, err := nats.Connect(natsUrl)
	if err != nil {
		log.WithError(err).Error("Failed to connect to nats")
		return
	}

	log.Info("Connected to nats at ", natsUrl)
	conn = c
}

func Publish(subject string, data []byte) {
	err := publishWithBreaker(subject, data)
	if err != nil {
		log.WithError(err).Error("Failed to publish message to nats at ", url)
	}
}

func publishWithBreaker(subject string, data []byte) error {
	defer func() error {
		if r := recover(); r != nil {
			return r.(error)
		}
		return nil
	}()

	_, err := circuitbreaker.NatsBreaker.Execute(func() (interface{}, error) {
		if conn == nil {
			log.Info("Trying to reconnect to nats at ", url)
			c, err := nats.Connect(url)
			if err != nil {
				return nil, err
			}
			log.Info("Reconnected to nats at ", url)
			conn = c
		}

		if !conn.IsConnected() {
			return nil, nats.ErrNoServers
		}

		err := conn.Publish(subject, data)
		if err != nil {
			return nil, err
		}

		err = conn.Flush()
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}
