package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"math"
)

type Royalty DefaultHandler

func newRoyaltyHandler(h *Handler) BuildInterface {
	return &Royalty{Handler: h}
}

func (h *Royalty) Validate() error {
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

func (h *Royalty) Build() (interface{}, error) {
	params, _ := h.GetParams()
	royalty, err := h.royaltyRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	merchant, err := h.merchantRepository.GetById(royalty.MerchantId.Hex())

	if err != nil {
		return nil, err
	}

	var products []map[string]interface{}
	var corrections []map[string]interface{}
	var summaryTotalEndUserSales int32
	var summaryTotalEndUserFees float64
	var summaryReturnsQty int32
	var summaryReturnsAmount float64
	var summarySalesCount int32
	var summaryEndUserFees float64
	var summaryVatOnEndUserSales float64
	var summaryLicenseRevenueShare float64
	var summaryLicenseFee float64

	for _, product := range royalty.Summary.ProductsItems {
		totalEndUserFees := math.Round(product.GrossSalesAmount*100) / 100
		returnsAmount := math.Round(product.GrossReturnsAmount*100) / 100
		endUserFees := math.Round(product.GrossTotalAmount*100) / 100
		vatOnEndUserSales := math.Round(product.TotalVat*100) / 100
		licenseRevenueShare := math.Round(product.TotalFees*100) / 100
		licenseFee := math.Round(product.PayoutAmount*100) / 100

		products = append(products, map[string]interface{}{
			"product":               product.Product,
			"region":                product.Region,
			"total_end_user_sales":  product.TotalTransactions,
			"total_end_user_fees":   totalEndUserFees,
			"returns_qty":           product.ReturnsCount,
			"returns_amount":        returnsAmount,
			"end_user_sales":        product.SalesCount,
			"end_user_fees":         endUserFees,
			"vat_on_end_user_sales": vatOnEndUserSales,
			"license_revenue_share": licenseRevenueShare,
			"license_fee":           licenseFee,
		})

		summaryTotalEndUserSales += product.TotalTransactions
		summaryTotalEndUserFees += totalEndUserFees
		summaryReturnsQty += product.ReturnsCount
		summaryReturnsAmount += returnsAmount
		summarySalesCount += product.SalesCount
		summaryEndUserFees += endUserFees
		summaryVatOnEndUserSales += vatOnEndUserSales
		summaryLicenseRevenueShare += licenseRevenueShare
		summaryLicenseFee += licenseFee
	}

	if len(royalty.Summary.Corrections) > 0 {
		for _, correction := range royalty.Summary.Corrections {
			t, err := ptypes.Timestamp(correction.EntryDate)

			if err != nil {
				return nil, err
			}

			corrections = append(corrections, map[string]interface{}{
				"entry_date": t.Format("2006-01-02T15:04:05"),
				"amount":     correction.Amount,
				"reason":     correction.Reason,
			})
		}
	}

	result := map[string]interface{}{
		"id":                       royalty.Id.Hex(),
		"report_date":              royalty.CreatedAt.Format("2006-01-02T15:04:05"),
		"merchant_legal_name":      merchant.Company.Name,
		"merchant_company_address": merchant.Company.Address,
		"start_date":               royalty.PeriodFrom.Format("2006-01-02T15:04:05"),
		"end_date":                 royalty.PeriodTo.Format("2006-01-02T15:04:05"),
		"currency":                 royalty.Currency,
		"correction_total_amount":  royalty.Totals.CorrectionAmount,
		"rolling_reserve_amount":   royalty.Totals.RollingReserveAmount,
		"products":                 products,
		"products_total": map[string]interface{}{
			"total_end_user_sales":  summaryTotalEndUserSales,
			"total_end_user_fees":   summaryTotalEndUserFees,
			"returns_qty":           summaryReturnsQty,
			"returns_amount":        summaryReturnsAmount,
			"end_user_sales":        summarySalesCount,
			"end_user_fees":         summaryEndUserFees,
			"vat_on_end_user_sales": summaryVatOnEndUserSales,
			"license_revenue_share": summaryLicenseRevenueShare,
			"license_fee":           summaryLicenseFee,
		},
		"corrections": corrections,
	}

	return result, nil
}

func (h *Royalty) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
