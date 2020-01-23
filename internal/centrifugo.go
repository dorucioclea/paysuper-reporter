package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/centrifugal/gocent"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

type CentrifugoInterface interface {
	Publish(string, interface{}) error
	Info(ctx context.Context) (gocent.InfoResult, error)
}

type Centrifugo struct {
	centrifugoClient *gocent.Client
}

type centrifugoHttpTransport struct {
	Transport http.RoundTripper
}

type centrifugoContextKey struct {
	name string
}

func newCentrifugoClient(cfg *config.CentrifugoConfig) CentrifugoInterface {
	return &Centrifugo{
		centrifugoClient: gocent.New(
			gocent.Config{
				Addr:       cfg.URL,
				Key:        cfg.ApiSecret,
				HTTPClient: &http.Client{Transport: &centrifugoHttpTransport{}},
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

func (c *Centrifugo) Info(ctx context.Context) (gocent.InfoResult, error) {
	return c.centrifugoClient.Info(ctx)
}

func (m *centrifugoHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), &centrifugoContextKey{name: "CentrifugoRequestStart"}, time.Now())
	req = req.WithContext(ctx)

	var reqBody []byte

	if req.Body != nil {
		reqBody, _ = ioutil.ReadAll(req.Body)
	}

	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))
	rsp, err := http.DefaultTransport.RoundTrip(req)

	if err != nil {
		return rsp, err
	}

	var rspBody []byte

	if rsp.Body != nil {
		rspBody, err = ioutil.ReadAll(rsp.Body)

		if err != nil {
			return rsp, err
		}
	}

	rsp.Body = ioutil.NopCloser(bytes.NewBuffer(rspBody))

	reqDecoded := make(map[string]interface{})
	err = json.Unmarshal(reqBody, &reqDecoded)

	if err != nil {
		return rsp, err
	}

	val, ok := reqDecoded["method"]

	if ok && val.(string) == "info" {
		return rsp, nil
	}

	req.Header.Set("Authorization", " ****** ")

	zap.L().Info(
		req.URL.Path,
		zap.Any("request_headers", req.Header),
		zap.ByteString("request_body", reqBody),
		zap.Int("response_status", rsp.StatusCode),
		zap.Any("response_headers", rsp.Header),
		zap.ByteString("response_body", rspBody),
	)

	return rsp, err
}
