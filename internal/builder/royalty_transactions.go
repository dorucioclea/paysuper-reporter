package builder

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
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

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *RoyaltyTransactions) Build() (interface{}, error) {
	royalty, err := h.royaltyReportRepository.GetById(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	orders, err := h.transactionsRepository.GetByRoyalty(royalty)

	if err != nil {
		return nil, err
	}

	return orders, nil
}
