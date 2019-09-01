package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/paysuper/paysuper-recurring-repository/tools"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
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
		httpClient: tools.NewLoggedHttpClient(zap.S()),
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
