package builder

import (
	"errors"
	"fmt"
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
	vat, err := h.vatReportRepository.GetById(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	orders, err := h.transactionsRepository.GetByVat(vat)

	if err != nil {
		return nil, err
	}

	return orders, nil
}
