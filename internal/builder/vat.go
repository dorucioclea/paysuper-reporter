package builder

import (
	"errors"
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
	royalty, err := h.vatReportRepository.GetById(h.report.Params[pkg.ParamsFieldId].(string))

	if err != nil {
		return nil, err
	}

	orders, err := h.vatReportRepository.GetTransactions(royalty)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"royalty": royalty, "orders": orders}, nil
}
