package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg"
	billingGrpc "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
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
	params, _ := h.GetParams()
	payout, err := h.payoutRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	merchant, err := h.merchantRepository.GetById(payout.MerchantId.Hex())

	if err != nil {
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
	}

	return result, nil
}

func (h *Payout) PostProcess(ctx context.Context, id string, fileName string, retentionTime int, content []byte) error {
	params, _ := h.GetParams()
	billingService := billingGrpc.NewBillingService(billingProto.ServiceName, h.service.Client())

	req := &billingGrpc.PayoutDocumentPdfUploadedRequest{
		Id:            id,
		PayoutId:      fmt.Sprintf("%s", params[pkg.ParamsFieldId]),
		Filename:      fileName,
		RetentionTime: int32(retentionTime),
		Content:       content,
	}

	if _, err := billingService.PayoutDocumentPdfUploaded(ctx, req); err != nil {
		return err
	}

	return nil
}
