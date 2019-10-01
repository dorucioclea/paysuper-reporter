package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
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
	params, _ := h.GetParams()
	country := fmt.Sprintf("%s", params[pkg.ParamsFieldCountry])
	vats, err := h.vatRepository.GetByCountry(country)

	if err != nil {
		return nil, err
	}

	grossRevenue := float64(0)
	correction := float64(0)
	totalTransactionsCount := int32(0)
	deduction := float64(0)
	ratesAndFees := float64(0)
	taxAmount := float64(0)

	var reports []map[string]interface{}

	for _, vat := range vats {
		grossRevenue += math.Round(vat.GrossRevenue*100) / 100
		correction += math.Round(vat.CorrectionAmount*100) / 100
		totalTransactionsCount += vat.TransactionsCount
		deduction += math.Round(vat.DeductionAmount*100) / 100
		ratesAndFees += math.Round(vat.FeesAmount*100) / 100
		taxAmount += math.Round(vat.VatAmount*100) / 100

		reports = append(reports, map[string]interface{}{
			"period_from":  vat.DateFrom.Format("2006-01-02T15:04:05"),
			"period_to":    vat.DateTo.Format("2006-01-02T15:04:05"),
			"report_date":  vat.CreatedAt.Format("2006-01-02T15:04:05"),
			"vat_id":       vat.Id.Hex(),
			"status":       vat.Status,
			"payment_date": vat.PayUntilDate.Format("2006-01-02T15:04:05"),
			"tax_amount":   math.Round(vat.VatAmount*100) / 100,
		})
	}

	result := map[string]interface{}{
		"country":                  country,
		"start_date":               "2019-10-01T00:00:00",
		"end_date":                 time.Now().Format("2006-01-02T15:04:05"),
		"gross_revenue":            grossRevenue,
		"correction":               correction,
		"total_transactions_count": totalTransactionsCount,
		"deduction":                deduction,
		"rates_and_fees":           ratesAndFees,
		"tax_amount":               taxAmount,
		"reports":                  reports,
	}

	return result, nil
}

func (h *Vat) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
