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

type Royalty DefaultHandler

func newRoyaltyHandler(h *Handler) BuildInterface {
	return &Royalty{Handler: h}
}

func (h *Royalty) Validate() error {
	params, err := h.GetParams()

	if err != nil {
		return err
	}

	if bson.IsObjectIdHex(h.report.MerchantId) != true {
		return errors.New(errs.ErrorParamMerchantIdNotFound.Message)
	}

	if _, ok := params[reporterpb.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", params[reporterpb.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Royalty) Build() (interface{}, error) {
	ctx := context.TODO()
	params, _ := h.GetParams()
	royaltyId := fmt.Sprintf("%s", params[reporterpb.ParamsFieldId])

	royaltyRequest := &billingpb.GetRoyaltyReportRequest{ReportId: royaltyId, MerchantId: h.report.MerchantId}
	royalty, err := h.billing.GetRoyaltyReport(ctx, royaltyRequest)

	if err != nil || royalty.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(royalty.Message.Message)
		}

		zap.L().Error(
			"Unable to get royalty report",
			zap.Error(err),
			zap.String("royalty_id", royaltyId),
		)

		return nil, err
	}

	merchantRequest := &billingpb.GetMerchantByRequest{MerchantId: h.report.MerchantId}
	merchant, err := h.billing.GetMerchantBy(ctx, merchantRequest)

	if err != nil || merchant.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(merchant.Message.Message)
		}

		zap.L().Error(
			"Unable to get merchant",
			zap.Error(err),
			zap.String("merchant_id", h.report.MerchantId),
		)

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

	for _, product := range royalty.Item.Summary.ProductsItems {
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

	if len(royalty.Item.Summary.Corrections) > 0 {
		for _, correction := range royalty.Item.Summary.Corrections {
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

	ocRequest := &billingpb.GetOperatingCompanyRequest{Id: royalty.Item.OperatingCompanyId}
	operatingCompany, err := h.billing.GetOperatingCompany(ctx, ocRequest)

	if err != nil || operatingCompany.Company == nil {
		if err == nil {
			err = errors.New(operatingCompany.Message.Message)
		}

		zap.L().Error(
			"Unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", royalty.Item.OperatingCompanyId),
		)

		return nil, err
	}

	date, err := ptypes.Timestamp(royalty.Item.CreatedAt)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("created_at", royalty.Item.CreatedAt.String()),
		)
		return nil, err
	}

	periodFrom, err := ptypes.Timestamp(royalty.Item.PeriodFrom)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("period_from", royalty.Item.PeriodFrom.String()),
		)
		return nil, err
	}

	periodTo, err := ptypes.Timestamp(royalty.Item.PeriodTo)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("period_to", royalty.Item.PeriodTo.String()),
		)
		return nil, err
	}

	result := map[string]interface{}{
		"id":                       royalty.Item.Id,
		"report_date":              date.Format("2006-01-02"),
		"merchant_legal_name":      merchant.Item.Company.Name,
		"merchant_company_address": merchant.Item.Company.Address,
		"start_date":               periodFrom.Format("2006-01-02"),
		"end_date":                 periodTo.Format("2006-01-02"),
		"currency":                 royalty.Item.Currency,
		"correction_total_amount":  royalty.Item.Totals.CorrectionAmount,
		"rolling_reserve_amount":   royalty.Item.Totals.RollingReserveAmount,
		"oc_name":                  operatingCompany.Company.Name,
		"oc_address":               operatingCompany.Company.Address,
		"products":                 products,
		"products_total": map[string]interface{}{
			"total_end_user_sales":  summaryTotalEndUserSales,
			"total_end_user_fees":   math.Round(summaryTotalEndUserFees*100) / 100,
			"returns_qty":           summaryReturnsQty,
			"returns_amount":        math.Round(summaryReturnsAmount*100) / 100,
			"end_user_sales":        summarySalesCount,
			"end_user_fees":         math.Round(summaryEndUserFees*100) / 100,
			"vat_on_end_user_sales": math.Round(summaryVatOnEndUserSales*100) / 100,
			"license_revenue_share": math.Round(summaryLicenseRevenueShare*100) / 100,
			"license_fee":           math.Round(summaryLicenseFee*100) / 100,
		},
		"corrections":     corrections,
		"has_corrections": len(corrections) > 0,
	}

	return result, nil
}

func (h *Royalty) PostProcess(
	ctx context.Context,
	id string,
	fileName string,
	retentionTime int64,
	content []byte,
) error {
	params, _ := h.GetParams()

	req := &billingpb.RoyaltyReportPdfUploadedRequest{
		Id:              id,
		RoyaltyReportId: fmt.Sprintf("%s", params[reporterpb.ParamsFieldId]),
		Filename:        fileName,
		RetentionTime:   int32(retentionTime),
		Content:         content,
	}

	if _, err := h.billing.RoyaltyReportPdfUploaded(ctx, req); err != nil {
		return err
	}

	return nil
}
