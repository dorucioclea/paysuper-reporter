package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
	"math"
)

type Payout DefaultHandler

func newPayoutHandler(h *Handler) BuildInterface {
	return &Payout{Handler: h}
}

func (h *Payout) Validate() error {
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

func (h *Payout) Build() (interface{}, error) {
	ctx := context.TODO()
	params, _ := h.GetParams()
	payoutId := fmt.Sprintf("%s", params[pkg.ParamsFieldId])

	payoutRequest := &billingpb.GetPayoutDocumentRequest{PayoutDocumentId: payoutId}
	payout, err := h.billing.GetPayoutDocument(ctx, payoutRequest)

	if err != nil || payout.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(payout.Message.Message)
		}

		zap.L().Error(
			"Unable to get payout document",
			zap.Error(err),
			zap.String("payout_id", payoutId),
		)

		return nil, err
	}

	merchantRequest := &billingpb.GetMerchantByRequest{MerchantId: payout.Item.MerchantId}
	merchant, err := h.billing.GetMerchantBy(ctx, merchantRequest)

	if err != nil || merchant.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(merchant.Message.Message)
		}

		zap.L().Error(
			"Unable to get merchant",
			zap.Error(err),
			zap.String("merchant_id", payout.Item.MerchantId),
		)

		return nil, err
	}

	ocRequest := &billingpb.GetOperatingCompanyRequest{Id: payout.Item.OperatingCompanyId}
	operatingCompany, err := h.billing.GetOperatingCompany(ctx, ocRequest)

	if err != nil || operatingCompany.Company == nil {
		if err == nil {
			err = errors.New(operatingCompany.Message.Message)
		}

		zap.L().Error(
			"Unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", payout.Item.OperatingCompanyId),
		)

		return nil, err
	}

	date, err := ptypes.Timestamp(payout.Item.CreatedAt)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("created_at", payout.Item.CreatedAt.String()),
		)
		return nil, err
	}

	periodFrom, err := ptypes.Timestamp(payout.Item.PeriodFrom)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("period_from", payout.Item.PeriodFrom.String()),
		)
		return nil, err
	}

	periodTo, err := ptypes.Timestamp(payout.Item.PeriodTo)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("period_to", payout.Item.PeriodTo.String()),
		)
		return nil, err
	}

	result := map[string]interface{}{
		"id":                      payout.Item.Id,
		"date":                    date.Format("2006-01-02"),
		"merchant_legal_name":     merchant.Item.Company.Name,
		"merchant_address":        merchant.Item.Company.Address,
		"merchant_eu_vat_number":  merchant.Item.Company.TaxId,
		"merchant_bank_details":   payout.Item.Destination.Details,
		"period_from":             periodFrom.Format("2006-01-02"),
		"period_to":               periodTo.Format("2006-01-02"),
		"transactions_for_period": payout.Item.TotalTransactions,
		"agreement_number":        payout.Item.MerchantAgreementNumber,
		"total_fees":              math.Round(payout.Item.TotalFees*100) / 100,
		"balance":                 math.Round(payout.Item.Balance*100) / 100,
		"currency":                payout.Item.Currency,
		"oc_name":                 operatingCompany.Company.Name,
		"oc_address":              operatingCompany.Company.Address,
		"oc_vat_number":           operatingCompany.Company.VatNumber,
		"oc_vat_address":          operatingCompany.Company.VatAddress,
	}

	return result, nil
}

func (h *Payout) PostProcess(
	ctx context.Context,
	id string,
	fileName string,
	retentionTime int64,
	content []byte,
) error {
	params, _ := h.GetParams()

	req := &billingpb.PayoutDocumentPdfUploadedRequest{
		Id:            id,
		PayoutId:      fmt.Sprintf("%s", params[pkg.ParamsFieldId]),
		Filename:      fileName,
		RetentionTime: int32(retentionTime),
		Content:       content,
	}

	if _, err := h.billing.PayoutDocumentPdfUploaded(ctx, req); err != nil {
		return err
	}

	return nil
}
