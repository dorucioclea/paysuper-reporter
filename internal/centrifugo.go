package internal

import (
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/paysuper/paysuper-reporter/internal/config"
	tools "github.com/paysuper/paysuper-tools/http"
	"go.uber.org/zap"
)

type CentrifugoInterface interface {
	Publish(string, interface{}) error
}

type Centrifugo struct {
	centrifugoClient *gocent.Client
}

func newCentrifugoClient(cfg *config.CentrifugoConfig) CentrifugoInterface {
	return &Centrifugo{
		centrifugoClient: gocent.New(
			gocent.Config{
				Addr:       cfg.URL,
				Key:        cfg.ApiSecret,
				HTTPClient: tools.NewLoggedHttpClient(zap.S()),
			},
		)}
}

func (c Centrifugo) Publish(channel string, msg interface{}) error {
	b, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	return c.centrifugoClient.Publish(context.Background(), channel, b)
}
