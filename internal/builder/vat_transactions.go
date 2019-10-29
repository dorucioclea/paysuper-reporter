package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
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
		amount := float64(0)
		amountCurrency := ""

		if order.PaymentGrossRevenueOrigin != nil {
			amount = order.PaymentGrossRevenueOrigin.Amount
			amountCurrency = order.PaymentGrossRevenueOrigin.Currency

			if order.Type == billingPkg.OrderTypeRefund {
				amount = -1 * order.PaymentRefundGrossRevenueOrigin.Amount
				amountCurrency = order.PaymentRefundGrossRevenueOrigin.Currency
			}
		}

		vat := float64(0)
		vatCurrency := ""

		if order.PaymentTaxFeeLocal != nil {
			vat = order.PaymentTaxFeeLocal.Amount
			vatCurrency = order.PaymentTaxFeeLocal.Currency

			if order.Type == billingPkg.OrderTypeRefund {
				vat = -1 * order.PaymentRefundTaxFeeLocal.Amount
				vatCurrency = order.PaymentRefundTaxFeeLocal.Currency
			}
		}

		fee := float64(0)
		feeCurrency := ""

		if order.FeesTotalLocal != nil {
			fee = order.FeesTotalLocal.Amount
			feeCurrency = order.FeesTotalLocal.Currency

			if order.Type == billingPkg.OrderTypeRefund {
				vat = order.RefundFeesTotalLocal.Amount
				vatCurrency = order.RefundFeesTotalLocal.Currency
			}
		}

		payout := float64(0)
		payoutCurrency := ""

		if order.NetRevenue != nil {
			payout = order.NetRevenue.Amount
			payoutCurrency = order.NetRevenue.Currency

			if order.Type == billingPkg.OrderTypeRefund {
				zap.L().Error(
					"debug refund payout",
					zap.String("id", order.Id.Hex()),
					zap.String("uuid", order.Uuid),
					zap.Float64("refund_reverse_revenue", order.RefundReverseRevenue.Amount),
				)
				vat = -1 * order.RefundReverseRevenue.Amount
				vatCurrency = order.RefundReverseRevenue.Currency
			}
		}

		isVatDeduction := "Yes"
		if !order.IsVatDeduction {
			isVatDeduction = "No"
		}

		transactions = append(transactions, map[string]interface{}{
			"date":             order.TransactionDate.Format("2006-01-02T15:04:05"),
			"country":          order.CountryCode,
			"id":               order.Id.Hex(),
			"payment_method":   order.PaymentMethod.Name,
			"amount":           math.Round(amount*100) / 100,
			"amount_currency":  amountCurrency,
			"vat":              math.Round(vat*100) / 100,
			"vat_currency":     vatCurrency,
			"fee":              math.Round(fee*100) / 100,
			"fee_currency":     feeCurrency,
			"payout":           math.Round(payout*100) / 100,
			"payout_currency":  payoutCurrency,
			"is_vat_deduction": isVatDeduction,
		})
	}

	result := map[string]interface{}{
		"id":                       params[pkg.ParamsFieldId],
		"country":                  vat.Country,
		"currency":                 vat.Currency,
		"vat_rate":                 vat.VatRate,
		"status":                   vat.Status,
		"pay_until_date":           vat.PayUntilDate,
		"country_annual_turnover":  vat.CountryAnnualTurnover,
		"world_annual_turnover":    vat.WorldAnnualTurnover,
		"created_at":               vat.CreatedAt.Format("2006-01-02"),
		"start_date":               vat.DateFrom.Format("2006-01-02"),
		"end_date":                 vat.DateTo.Format("2006-01-02"),
		"gross_revenue":            math.Round(vat.GrossRevenue*100) / 100,
		"correction":               math.Round(vat.CorrectionAmount*100) / 100,
		"total_transactions_count": vat.TransactionsCount,
		"deduction":                math.Round(vat.DeductionAmount*100) / 100,
		"rates_and_fees":           math.Round(vat.FeesAmount*100) / 100,
		"tax_amount":               math.Round(vat.VatAmount*100) / 100,
		"has_pay_until_date":       vat.Status == billingPkg.VatReportStatusNeedToPay || vat.Status == billingPkg.VatReportStatusOverdue,
		"has_disclaimer":           vat.AmountsApproximate,
		"transactions":             transactions,
	}

	return result, nil
}

func (h *VatTransactions) PostProcess(ctx context.Context, id string, fileName string, retentionTime int, content []byte) error {
	return nil
}
