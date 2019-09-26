package builder

import (
	"context"
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

func (h *RoyaltyTransactions) Build() (interface{}, error) {
	params, _ := h.GetParams()
	royalty, err := h.royaltyRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	orders, err := h.transactionsRepository.GetByRoyalty(royalty)

	if err != nil {
		return nil, err
	}

	var transactions []map[string]interface{}

	for _, order := range orders {
		transactions = append(transactions, map[string]interface{}{
			"transaction":                   order.Transaction,
			"country_code":                  order.CountryCode,
			"total_payment_amount":          order.TotalPaymentAmount,
			"currency":                      order.Currency,
			"payment_method":                order.PaymentMethod.Name,
			"created_at":                    order.CreatedAt.Format("2006-01-02T15:04:05"),
			"gross_revenue":                 order.GrossRevenue.Amount,
			"tax_fee":                       order.TaxFee.Amount,
			"tax_fee_currency_exchange_fee": order.TaxFeeCurrencyExchangeFee.Amount,
			"tax_fee_total":                 order.TaxFeeTotal.Amount,
			"method_fee_total":              order.MethodFeeTotal.Amount,
			"method_fee_tariff":             order.MethodFeeTariff.Amount,
			"method_fixed_fee_tariff":       order.MethodFixedFeeTariff.Amount,
			"paysuper_fixed_fee":            order.PaysuperFixedFee.Amount,
			"fees_total":                    order.FeesTotal.Amount,
			"fees_total_local":              order.FeesTotalLocal.Amount,
			"net_revenue":                   order.NetRevenue.Amount,
		})
	}

	result := map[string]interface{}{
		"id":           params[pkg.ParamsFieldId],
		"transactions": transactions,
	}

	return result, nil
}

func (h *RoyaltyTransactions) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
