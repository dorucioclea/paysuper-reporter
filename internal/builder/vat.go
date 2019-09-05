package builder

import (
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type Vat DefaultHandler

func newVatHandler(h *Handler) BuildInterface {
	return &Vat{Handler: h}
}

func (h *Vat) Validate() error {
	if _, ok := h.report.Params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Vat) Build() (interface{}, error) {
	// TODO: Remove me!
	return map[string]interface{}{
		"id":   1,
		"name": "test",
	}, nil

	vat, err := h.vatReportRepository.GetById(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	return vat, nil
}
