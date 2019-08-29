package builder

import (
	"errors"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type Royalty DefaultHandler

func newRoyaltyHandler(h *Handler) BuildInterface {
	return &Royalty{Handler: h}
}

func (h *Royalty) Validate() error {
	if _, ok := h.report.Params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Royalty) Build() (interface{}, error) {
	royalty, err := h.royaltyReportRepository.GetById(h.report.Params[pkg.ParamsFieldId].(string))

	if err != nil {
		return nil, err
	}

	orders, err := h.royaltyReportRepository.GetTransactions(royalty)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"royalty": royalty, "orders": orders}, nil
}
