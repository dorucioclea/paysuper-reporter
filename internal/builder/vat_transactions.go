package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"math"
)

type VatTransactions DefaultHandler

func newVatTransactionsHandler(h *Handler) BuildInterface {
	return &VatTransactions{Handler: h}
}

func (h *VatTransactions) Validate() error {
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

func (h *VatTransactions) Build() (interface{}, error) {
	params, _ := h.GetParams()
	vat, err := h.vatRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	orders, err := h.transactionsRepository.GetByVat(vat)

	if err != nil {
		return nil, err
	}

	var transactions []map[string]interface{}

	for _, order := range orders {
		taxFeeTotal := float64(0)
		if order.TaxFeeTotal != nil {
			taxFeeTotal = order.TaxFeeTotal.Amount
		}

		feesTotal := float64(0)
		if order.FeesTotal != nil {
			feesTotal = order.FeesTotal.Amount
		}

		grossRevenue := float64(0)
		if order.GrossRevenue != nil {
			grossRevenue = order.GrossRevenue.Amount
		}

		transactions = append(transactions, map[string]interface{}{
			"date":           order.TransactionDate.Format("2006-01-02T15:04:05"),
			"country":        order.CountryCode,
			"id":             order.Id.Hex(),
			"payment_method": order.PaymentMethod.Name,
			"amount":         math.Round(order.TotalPaymentAmount*100) / 100,
			"vat":            math.Round(taxFeeTotal*100) / 100,
			"fee":            math.Round(feesTotal*100) / 100,
			"payout":         math.Round(grossRevenue*100) / 100,
		})
	}

	result := map[string]interface{}{
		"id":                       params[pkg.ParamsFieldId],
		"country":                  vat.Country,
		"created_at":               vat.CreatedAt.Format("2006-01-02T15:04:05"),
		"start_date":               vat.DateFrom.Format("2006-01-02T15:04:05"),
		"end_date":                 vat.DateTo.Format("2006-01-02T15:04:05"),
		"gross_revenue":            math.Round(vat.GrossRevenue*100) / 100,
		"correction":               math.Round(vat.CorrectionAmount*100) / 100,
		"total_transactions_count": vat.TransactionsCount,
		"deduction":                math.Round(vat.DeductionAmount*100) / 100,
		"rates_and_fees":           math.Round(vat.FeesAmount*100) / 100,
		"tax_amount":               math.Round(vat.VatAmount*100) / 100,
		"transactions":             transactions,
	}

	return result, nil
}

func (h *VatTransactions) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
