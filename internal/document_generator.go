package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-recurring-repository/tools"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type DocumentGeneratorInterface interface {
	Render(payload *proto.GeneratorPayload) ([]byte, error)
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

func newDocumentGenerator(config *config.DocumentGeneratorConfig) DocumentGeneratorInterface {
	client := &DocumentGenerator{
		apiUrl:     config.ApiUrl,
		timeout:    config.Timeout,
		httpClient: tools.NewLoggedHttpClient(zap.S()),
	}

	return client
}

func (dg DocumentGenerator) Render(payload *proto.GeneratorPayload) ([]byte, error) {
	b, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	rsp, err := dg.httpClient.Post(dg.apiUrl+"/api/report", pkg.MIMEApplicationJSON, bytes.NewBuffer(b))

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	var msg []byte
	msg, err = ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != 200 {
		var rspErr map[string]interface{}

		if err = json.Unmarshal(b, &rspErr); err != nil {
			return nil, errors.New(errs.ErrorDocumentGeneratorRender.Message)
		}

		return nil, errors.New(fmt.Sprintf("%s", rspErr["message"]))
	}

	return msg, nil
}
