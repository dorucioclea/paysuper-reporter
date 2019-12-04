package builder

import (
	"context"
	"errors"
	"fmt"
	billingGrpc "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
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

	if _, ok := params[pkg.ParamsFieldCountry]; !ok {
		return errors.New(errs.ErrorParamCountryNotFound.Message)
	}

	if len(fmt.Sprintf("%s", params[pkg.ParamsFieldCountry])) != 2 {
		return errors.New(errs.ErrorParamCountryNotFound.Message)
	}

	return nil
}

func (h *Vat) Build() (interface{}, error) {
	var reports []map[string]interface{}

	params, _ := h.GetParams()
	country := fmt.Sprintf("%s", params[pkg.ParamsFieldCountry])
	vats, err := h.vatRepository.GetByCountry(country)

	if err != nil {
		return nil, err
	}

	if len(vats) < 1 {
		return reports, nil
	}

	grossRevenue := float64(0)
	correction := float64(0)
	totalTransactionsCount := int32(0)
	deduction := float64(0)
	ratesAndFees := float64(0)
	taxAmount := float64(0)

	for _, vat := range vats {
		grossRevenue += math.Round(vat.GrossRevenue*100) / 100
		correction += math.Round(vat.CorrectionAmount*100) / 100
		totalTransactionsCount += vat.TransactionsCount
		deduction += math.Round(vat.DeductionAmount*100) / 100
		ratesAndFees += math.Round(vat.FeesAmount*100) / 100
		taxAmount += math.Round(vat.VatAmount*100) / 100

		reports = append(reports, map[string]interface{}{
			"period_from":             vat.DateFrom.Format("2006-01-02"),
			"period_to":               vat.DateTo.Format("2006-01-02"),
			"vat_id":                  vat.Id.Hex(),
			"status":                  vat.Status,
			"payment_date":            vat.PayUntilDate.Format("2006-01-02"),
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
		&billingGrpc.GetOperatingCompanyRequest{Id: vats[0].OperatingCompanyId},
	)

	if err != nil || res.Company == nil {
		if err == nil {
			err = errors.New(res.Message.Message)
		}

		zap.L().Error(
			"unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", vats[0].OperatingCompanyId),
		)

		return nil, err
	}

	result := map[string]interface{}{
		"country":                  country,
		"currency":                 vats[0].Currency,
		"vat_rate":                 vats[0].VatRate,
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
