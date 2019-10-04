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
	username   string
	password   string
	httpClient *http.Client
}

func newDocumentGenerator(config *config.DocumentGeneratorConfig) DocumentGeneratorInterface {
	client := &DocumentGenerator{
		apiUrl:     config.ApiUrl,
		timeout:    config.Timeout,
		username:   config.Username,
		password:   config.Password,
		httpClient: tools.NewLoggedHttpClient(zap.S()),
	}

	return client
}

func (dg DocumentGenerator) Render(payload *proto.GeneratorPayload) ([]byte, error) {
	b, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", dg.apiUrl+"/api/report", bytes.NewBuffer(b))

	if err != nil {
		return nil, err
	}

	if dg.username != "" && dg.password != "" {
		req.SetBasicAuth(dg.username, dg.password)
	}

	req.Header.Set("Content-Type", pkg.MIMEApplicationJSON)
	rsp, err := dg.httpClient.Do(req)

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

		if err = json.Unmarshal(msg, &rspErr); err != nil {
			return nil, errors.New("error unmarshal jsreport response: " + errs.ErrorDocumentGeneratorRender.Message)
		}

		return nil, errors.New("error jsreport response code: " + string(msg))
	}

	return msg, nil
}
