package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
	"math"
	"time"
)

type Vat DefaultHandler

func newVatHandler(h *Handler) BuildInterface {
	return &Vat{Handler: h}
}

func (h *Vat) Validate() error {
	params, err := h.GetParams()

	if err != nil {
		return err
	}

	if _, ok := params[reporterpb.ParamsFieldCountry]; !ok {
		return errors.New(errs.ErrorParamCountryNotFound.Message)
	}

	if len(fmt.Sprintf("%s", params[reporterpb.ParamsFieldCountry])) != 2 {
		return errors.New(errs.ErrorParamCountryNotFound.Message)
	}

	return nil
}

func (h *Vat) Build() (interface{}, error) {
	var reports []map[string]interface{}

	ctx := context.TODO()
	params, _ := h.GetParams()
	country := fmt.Sprintf("%s", params[reporterpb.ParamsFieldCountry])

	vatsRequest := &billingpb.VatReportsRequest{Country: country, Offset: 0, Limit: 1000}
	vats, err := h.billing.GetVatReportsForCountry(ctx, vatsRequest)

	if err != nil || vats.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(vats.Message.Message)
		}

		zap.L().Error(
			"Unable to get vats for country",
			zap.Error(err),
			zap.String("country", country),
		)

		return nil, err
	}

	if len(vats.Data.Items) < 1 {
		return reports, nil
	}

	grossRevenue := float64(0)
	correction := float64(0)
	totalTransactionsCount := int32(0)
	deduction := float64(0)
	ratesAndFees := float64(0)
	taxAmount := float64(0)

	for _, vat := range vats.Data.Items {
		grossRevenue += math.Round(vat.GrossRevenue*100) / 100
		correction += math.Round(vat.CorrectionAmount*100) / 100
		totalTransactionsCount += vat.TransactionsCount
		deduction += math.Round(vat.DeductionAmount*100) / 100
		ratesAndFees += math.Round(vat.FeesAmount*100) / 100
		taxAmount += math.Round(vat.VatAmount*100) / 100

		dateFrom, err := ptypes.Timestamp(vat.DateFrom)

		if err != nil {
			zap.L().Error(
				"Unable to cast timestamp to time",
				zap.Error(err),
				zap.String("date_from", vat.DateFrom.String()),
			)
			return nil, err
		}

		dateTo, err := ptypes.Timestamp(vat.DateTo)

		if err != nil {
			zap.L().Error(
				"Unable to cast timestamp to time",
				zap.Error(err),
				zap.String("date_to", vat.DateTo.String()),
			)
			return nil, err
		}

		payUntilDate, err := ptypes.Timestamp(vat.PayUntilDate)

		if err != nil {
			zap.L().Error(
				"Unable to cast timestamp to time",
				zap.Error(err),
				zap.String("pay_until_date", vat.PayUntilDate.String()),
			)
			return nil, err
		}

		reports = append(reports, map[string]interface{}{
			"period_from":             dateFrom.Format("2006-01-02"),
			"period_to":               dateTo.Format("2006-01-02"),
			"vat_id":                  vat.Id,
			"status":                  vat.Status,
			"payment_date":            payUntilDate.Format("2006-01-02"),
			"tax_amount":              math.Round(vat.VatAmount*100) / 100,
			"transactions_count":      vat.TransactionsCount,
			"gross_amount":            math.Round(vat.GrossRevenue*100) / 100,
			"deduction_amount":        math.Round(vat.DeductionAmount*100) / 100,
			"correction_amount":       math.Round(vat.CorrectionAmount*100) / 100,
			"country_annual_turnover": math.Round(vat.CountryAnnualTurnover*100) / 100,
			"world_annual_turnover":   math.Round(vat.WorldAnnualTurnover*100) / 100,
		})
	}

	res, err := h.billing.GetOperatingCompany(
		context.Background(),
		&billingpb.GetOperatingCompanyRequest{Id: vats.Data.Items[0].OperatingCompanyId},
	)

	if err != nil || res.Company == nil {
		if err == nil {
			err = errors.New(res.Message.Message)
		}

		zap.L().Error(
			"unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", vats.Data.Items[0].OperatingCompanyId),
		)

		return nil, err
	}

	result := map[string]interface{}{
		"country":                  country,
		"currency":                 vats.Data.Items[0].Currency,
		"vat_rate":                 vats.Data.Items[0].VatRate,
		"start_date":               "2019-10-01",
		"end_date":                 time.Now().Format("2006-01-02"),
		"gross_revenue":            grossRevenue,
		"correction":               correction,
		"total_transactions_count": totalTransactionsCount,
		"deduction":                deduction,
		"rates_and_fees":           ratesAndFees,
		"tax_amount":               taxAmount,
		"has_total_block":          len(reports) > 1,
		"oc_name":                  res.Company.Name,
		"oc_address":               res.Company.Address,
		"reports":                  reports,
	}

	return result, nil
}

func (h *Vat) PostProcess(ctx context.Context, id, fileName string, retentionTime int64, content []byte) error {
	return nil
}
