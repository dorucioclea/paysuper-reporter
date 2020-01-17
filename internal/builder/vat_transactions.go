package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
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

	if _, ok := params[reporterpb.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", params[reporterpb.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *VatTransactions) Build() (interface{}, error) {
	ctx := context.TODO()
	params, _ := h.GetParams()
	vatId := fmt.Sprintf("%s", params[reporterpb.ParamsFieldId])

	vatRequest := &billingpb.VatReportRequest{Id: vatId}
	vat, err := h.billing.GetVatReport(ctx, vatRequest)

	if err != nil || vat.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(vat.Message.Message)
		}

		zap.L().Error(
			"Unable to get vat orders",
			zap.Error(err),
			zap.String("vat_id", vatId),
		)

		return nil, err
	}

	ordersRequest := &billingpb.VatTransactionsRequest{VatReportId: vatId, Offset: 0, Limit: 1000}
	orders, err := h.billing.GetVatReportTransactions(ctx, ordersRequest)

	if err != nil || orders.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(orders.Message.Message)
		}

		zap.L().Error(
			"Unable to get vat orders",
			zap.Error(err),
			zap.String("vat_id", vatId),
		)

		return nil, err
	}

	var transactions []map[string]interface{}

	for _, order := range orders.Data.Items {
		amount := float64(0)
		amountCurrency := ""

		if order.PaymentGrossRevenueOrigin != nil {
			amount = order.PaymentGrossRevenueOrigin.Amount
			amountCurrency = order.PaymentGrossRevenueOrigin.Currency

			if order.Type == billingpb.OrderTypeRefund {
				amount = -1 * order.PaymentRefundGrossRevenueOrigin.Amount
				amountCurrency = order.PaymentRefundGrossRevenueOrigin.Currency
			}
		}

		vat := float64(0)
		vatCurrency := ""

		if order.PaymentTaxFeeLocal != nil {
			vat = order.PaymentTaxFeeLocal.Amount
			vatCurrency = order.PaymentTaxFeeLocal.Currency

			if order.Type == billingpb.OrderTypeRefund {
				vat = -1 * order.PaymentRefundTaxFeeLocal.Amount
				vatCurrency = order.PaymentRefundTaxFeeLocal.Currency
			}
		}

		fee := float64(0)
		feeCurrency := ""

		if order.FeesTotalLocal != nil {
			fee = order.FeesTotalLocal.Amount
			feeCurrency = order.FeesTotalLocal.Currency

			if order.Type == billingpb.OrderTypeRefund {
				vat = order.RefundFeesTotalLocal.Amount
				vatCurrency = order.RefundFeesTotalLocal.Currency
			}
		}

		payout := float64(0)
		payoutCurrency := ""

		if order.NetRevenue != nil {
			payout = order.NetRevenue.Amount
			payoutCurrency = order.NetRevenue.Currency

			if order.Type == billingpb.OrderTypeRefund {
				zap.L().Error(
					"debug refund payout",
					zap.String("id", order.Id),
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

		date, err := ptypes.Timestamp(order.TransactionDate)

		if err != nil {
			zap.L().Error(
				"Unable to cast timestamp to time",
				zap.Error(err),
				zap.String("transaction_date", order.TransactionDate.String()),
			)
			return nil, err
		}

		transactions = append(transactions, map[string]interface{}{
			"date":             date.Format("2006-01-02T15:04:05"),
			"country":          order.CountryCode,
			"id":               order.Id,
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

	res, err := h.billing.GetOperatingCompany(
		context.Background(),
		&billingpb.GetOperatingCompanyRequest{Id: vat.Vat.OperatingCompanyId},
	)

	if err != nil || res.Company == nil {
		if err == nil {
			err = errors.New(res.Message.Message)
		}

		zap.L().Error(
			"unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", vat.Vat.OperatingCompanyId),
		)

		return nil, err
	}

	createdAt, err := ptypes.Timestamp(vat.Vat.CreatedAt)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("created_at", vat.Vat.CreatedAt.String()),
		)
		return nil, err
	}

	dateFrom, err := ptypes.Timestamp(vat.Vat.DateFrom)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("date_from", vat.Vat.DateFrom.String()),
		)
		return nil, err
	}

	dateTo, err := ptypes.Timestamp(vat.Vat.DateTo)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("date_to", vat.Vat.DateTo.String()),
		)
		return nil, err
	}

	result := map[string]interface{}{
		"id":                       params[reporterpb.ParamsFieldId],
		"country":                  vat.Vat.Country,
		"currency":                 vat.Vat.Currency,
		"vat_rate":                 vat.Vat.VatRate,
		"status":                   vat.Vat.Status,
		"pay_until_date":           vat.Vat.PayUntilDate,
		"country_annual_turnover":  vat.Vat.CountryAnnualTurnover,
		"world_annual_turnover":    vat.Vat.WorldAnnualTurnover,
		"created_at":               createdAt.Format("2006-01-02"),
		"start_date":               dateFrom.Format("2006-01-02"),
		"end_date":                 dateTo.Format("2006-01-02"),
		"gross_revenue":            math.Round(vat.Vat.GrossRevenue*100) / 100,
		"correction":               math.Round(vat.Vat.CorrectionAmount*100) / 100,
		"total_transactions_count": vat.Vat.TransactionsCount,
		"deduction":                math.Round(vat.Vat.DeductionAmount*100) / 100,
		"rates_and_fees":           math.Round(vat.Vat.FeesAmount*100) / 100,
		"tax_amount":               math.Round(vat.Vat.VatAmount*100) / 100,
		"has_pay_until_date":       vat.Vat.Status == billingpb.VatReportStatusNeedToPay || vat.Vat.Status == billingpb.VatReportStatusOverdue,
		"has_disclaimer":           vat.Vat.AmountsApproximate,
		"oc_name":                  res.Company.Name,
		"oc_address":               res.Company.Address,
		"transactions":             transactions,
	}

	return result, nil
}

func (h *VatTransactions) PostProcess(ctx context.Context, id, fileName string, retentionTime int64, content []byte) error {
	return nil
}
