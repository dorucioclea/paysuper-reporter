package builder

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
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

	if _, ok := params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Vat) Build() (interface{}, error) {
	params, _ := h.GetParams()
	vat, err := h.vatReportRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"id":                      vat.Id.Hex(),
		"country":                 vat.Country,
		"vat_rate":                vat.VatRate,
		"currency":                vat.Currency,
		"transactions_count":      vat.TransactionsCount,
		"gross_revenue":           vat.GrossRevenue,
		"vat_amount":              vat.VatAmount,
		"fees_amount":             vat.FeesAmount,
		"deduction_amount":        vat.DeductionAmount,
		"correction_amount":       vat.CorrectionAmount,
		"country_annual_turnover": vat.CountryAnnualTurnover,
		"world_annual_turnover":   vat.WorldAnnualTurnover,
		"amounts_approximate":     vat.AmountsApproximate,
		"date_from":               vat.DateFrom.Format("2006-01-02T15:04:05"),
		"date_to":                 vat.DateTo.Format("2006-01-02T15:04:05"),
		"pay_until_date":          vat.PayUntilDate.Format("2006-01-02T15:04:05"),
		"created_at":              vat.CreatedAt.Format("2006-01-02T15:04:05"),
	}

	return result, nil
}
