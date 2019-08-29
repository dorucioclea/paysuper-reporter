package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	documentGeneratorContextKey = "DocumentGeneratorConfig"
)

type DocumentGeneratorInterface interface {
	Render(payload *proto.GeneratorPayload) (*proto.File, error)
}

type DocumentGeneratorRenderRequest struct {
	TemplateId string
	Data       interface{}
	Options    interface{}
}

type DocumentGenerator struct {
	apiUrl     string
	timeout    int
	httpClient *http.Client
}

func newDocumentGenerator(config *config.DocumentGeneratorConfig) (DocumentGeneratorInterface, error) {
	client := DocumentGenerator{
		apiUrl:     config.ApiUrl,
		timeout:    config.Timeout,
		httpClient: &http.Client{Transport: &httpTransport{}},
	}

	return client, nil
}

func (dg DocumentGenerator) Render(payload *proto.GeneratorPayload) (*proto.File, error) {
	b, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, dg.apiUrl, bytes.NewBuffer(b))

	if err != nil {
		return nil, err
	}

	req.Header.Add(pkg.HeaderContentType, pkg.MIMEApplicationJSON)
	req.Header.Add(pkg.HeaderAccept, pkg.MIMEApplicationJSON)

	rsp, err := dg.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	b, err = ioutil.ReadAll(rsp.Body)
	_ = rsp.Body.Close()

	if err != nil {
		return nil, err
	}

	msg := &proto.File{}
	err = json.Unmarshal(b, msg)

	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != http.StatusOK {
		return nil, errors.New(errs.ErrorDocumentGeneratorRender.Message)
	}

	return msg, nil
}

type httpTransport struct {
	Transport http.RoundTripper
}

func (t *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := context.WithValue(req.Context(), documentGeneratorContextKey, time.Now())
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
