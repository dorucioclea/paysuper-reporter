package builder

import (
	"errors"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type RoyaltyTransactions DefaultHandler

func newRoyaltyTransactionsHandler(h *Handler) BuildInterface {
	return &RoyaltyTransactions{Handler: h}
}

func (h *RoyaltyTransactions) Validate() error {
	if _, ok := h.report.Params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *RoyaltyTransactions) Build() (interface{}, error) {
	royalty, err := h.royaltyReportRepository.GetById(h.report.Params[pkg.ParamsFieldId].(string))

	if err != nil {
		return nil, err
	}

	orders, err := h.royaltyReportRepository.GetTransactions(royalty)

	if err != nil {
		return nil, err
	}

	return orders, nil
}
