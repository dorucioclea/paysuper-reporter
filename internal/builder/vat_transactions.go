package builder

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
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
	vat, err := h.vatReportRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	orders, err := h.transactionsRepository.GetByVat(vat)

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
