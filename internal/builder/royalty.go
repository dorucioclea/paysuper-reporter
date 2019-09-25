package builder

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type Royalty DefaultHandler

func newRoyaltyHandler(h *Handler) BuildInterface {
	return &Royalty{Handler: h}
}

func (h *Royalty) Validate() error {
	params, err := h.GetParams()

	if err != nil {
		return err
	}

	if _, ok := params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Royalty) Build() (interface{}, error) {
	params, _ := h.GetParams()
	royalty, err := h.royaltyReportRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":                         royalty.Id.Hex(),
		"payout_id":                  royalty.PayoutId,
		"correction":                 royalty.Correction.Amount,
		"period_from":                royalty.PeriodFrom.Format("2006-01-02T15:04:05"),
		"period_to":                  royalty.PeriodTo.Format("2006-01-02T15:04:05"),
		"payout_date":                royalty.PayoutDate.Format("2006-01-02T15:04:05"),
		"created_at":                 royalty.CreatedAt.Format("2006-01-02T15:04:05"),
		"accepted_at":                royalty.AcceptedAt.Format("2006-01-02T15:04:05"),
		"amounts_vat":                royalty.Amounts.VatAmount,
		"amounts_transactions_count": royalty.Amounts.TransactionsCount,
		"amounts_currency":           royalty.Amounts.Currency,
		"amounts_fee":                royalty.Amounts.FeeAmount,
		"amounts_gross":              royalty.Amounts.GrossAmount,
		"amounts_payout":             royalty.Amounts.PayoutAmount,
	}

	return result, nil
}
