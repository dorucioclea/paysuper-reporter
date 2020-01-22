package builder

import (
	"context"
	"errors"
	"fmt"
	billingGrpc "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	_, err = primitive.ObjectIDFromHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Payout) Build() (interface{}, error) {
	params, _ := h.GetParams()
	payout, err := h.payoutRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	merchant, err := h.merchantRepository.GetById(payout.MerchantId.Hex())

	if err != nil {
		return nil, err
	}

	res, err := h.billing.GetOperatingCompany(
		context.Background(),
		&billingGrpc.GetOperatingCompanyRequest{Id: payout.OperatingCompanyId},
	)

	if err != nil || res.Company == nil {
		if err == nil {
			err = errors.New(res.Message.Message)
		}

		zap.L().Error(
			"unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", payout.OperatingCompanyId),
		)

		return nil, err
	}

	result := map[string]interface{}{
		"id":                      payout.Id.Hex(),
		"date":                    payout.CreatedAt.Format("2006-01-02"),
		"merchant_legal_name":     merchant.Company.Name,
		"merchant_address":        merchant.Company.Address,
		"merchant_eu_vat_number":  merchant.Company.TaxId,
		"merchant_bank_details":   payout.Destination.Details,
		"period_from":             payout.PeriodFrom.Format("2006-01-02"),
		"period_to":               payout.PeriodTo.Format("2006-01-02"),
		"transactions_for_period": payout.TotalTransactions,
		"agreement_number":        payout.MerchantAgreementNumber,
		"total_fees":              math.Round(payout.TotalFees*100) / 100,
		"balance":                 math.Round(payout.Balance*100) / 100,
		"currency":                payout.Currency,
		"oc_name":                 res.Company.Name,
		"oc_address":              res.Company.Address,
		"oc_vat_number":           res.Company.VatNumber,
		"oc_vat_address":          res.Company.VatAddress,
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

	req := &billingGrpc.PayoutDocumentPdfUploadedRequest{
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
