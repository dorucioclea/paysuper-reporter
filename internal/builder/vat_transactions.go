package builder

import (
	"errors"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type VatTransactions DefaultHandler

func newVatTransactionsHandler(h *Handler) BuildInterface {
	return &VatTransactions{Handler: h}
}

func (h *VatTransactions) Validate() error {
	if _, ok := h.report.Params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *VatTransactions) Build() (interface{}, error) {
	vat, err := h.vatReportRepository.GetById(h.report.Params[pkg.ParamsFieldId].(string))

	if err != nil {
		return nil, err
	}

	orders, err := h.vatReportRepository.GetTransactions(vat)

	if err != nil {
		return nil, err
	}

	return orders, nil
}
