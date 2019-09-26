package builder

import (
	"context"
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
	royalty, err := h.royaltyRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":                         royalty.Id.Hex(),
		"period_from":                royalty.PeriodFrom.Format("2006-01-02T15:04:05"),
		"period_to":                  royalty.PeriodTo.Format("2006-01-02T15:04:05"),
		"payout_date":                royalty.PayoutDate.Format("2006-01-02T15:04:05"),
		"created_at":                 royalty.CreatedAt.Format("2006-01-02T15:04:05"),
		"accepted_at":                royalty.AcceptedAt.Format("2006-01-02T15:04:05"),
		"amounts_vat":                royalty.Totals.VatAmount,
		"amounts_transactions_count": royalty.Totals.TransactionsCount,
		"amounts_rolling_reserve":    royalty.Totals.RollingReserveAmount,
		"amounts_fee":                royalty.Totals.FeeAmount,
		"amounts_correction":         royalty.Totals.CorrectionAmount,
		"amounts_payout":             royalty.Totals.PayoutAmount,
	}

	return result, nil
}

func (h *Royalty) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
