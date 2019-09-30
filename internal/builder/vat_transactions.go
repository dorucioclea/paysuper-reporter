package builder

import (
	"context"
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
		grossRevenue := float64(0)
		if order.GrossRevenue != nil {
			grossRevenue = order.GrossRevenue.Amount
		}

		taxFee := float64(0)
		if order.TaxFee != nil {
			taxFee = order.TaxFee.Amount
		}

		taxFeeCurrencyExchangeFee := float64(0)
		if order.TaxFeeCurrencyExchangeFee != nil {
			taxFeeCurrencyExchangeFee = order.TaxFeeCurrencyExchangeFee.Amount
		}

		taxFeeTotal := float64(0)
		if order.TaxFeeTotal != nil {
			taxFeeTotal = order.TaxFeeTotal.Amount
		}

		methodFeeTotal := float64(0)
		if order.MethodFeeTotal != nil {
			methodFeeTotal = order.MethodFeeTotal.Amount
		}

		methodFeeTariff := float64(0)
		if order.MethodFeeTariff != nil {
			methodFeeTariff = order.MethodFeeTariff.Amount
		}

		methodFixedFeeTariff := float64(0)
		if order.MethodFixedFeeTariff != nil {
			methodFixedFeeTariff = order.MethodFixedFeeTariff.Amount
		}

		paysuperFixedFee := float64(0)
		if order.PaysuperFixedFee != nil {
			paysuperFixedFee = order.PaysuperFixedFee.Amount
		}

		feesTotal := float64(0)
		if order.FeesTotal != nil {
			feesTotal = order.FeesTotal.Amount
		}

		feesTotalLocal := float64(0)
		if order.FeesTotalLocal != nil {
			feesTotalLocal = order.FeesTotalLocal.Amount
		}

		netRevenue := float64(0)
		if order.NetRevenue != nil {
			netRevenue = order.NetRevenue.Amount
		}

		transactions = append(transactions, map[string]interface{}{
			"transaction":                   order.Transaction,
			"country_code":                  order.CountryCode,
			"total_payment_amount":          order.TotalPaymentAmount,
			"currency":                      order.Currency,
			"payment_method":                order.PaymentMethod.Name,
			"created_at":                    order.CreatedAt.Format("2006-01-02T15:04:05"),
			"gross_revenue":                 grossRevenue,
			"tax_fee":                       taxFee,
			"tax_fee_currency_exchange_fee": taxFeeCurrencyExchangeFee,
			"tax_fee_total":                 taxFeeTotal,
			"method_fee_total":              methodFeeTotal,
			"method_fee_tariff":             methodFeeTariff,
			"method_fixed_fee_tariff":       methodFixedFeeTariff,
			"paysuper_fixed_fee":            paysuperFixedFee,
			"fees_total":                    feesTotal,
			"fees_total_local":              feesTotalLocal,
			"net_revenue":                   netRevenue,
		})
	}

	result := map[string]interface{}{
		"id":           params[pkg.ParamsFieldId],
		"transactions": transactions,
	}

	return result, nil
}

func (h *VatTransactions) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
