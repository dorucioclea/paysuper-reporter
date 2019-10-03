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

	result := map[string]interface{}{
		"id":                     payout.Id.Hex(),
		"date":                   payout.CreatedAt.Format("2006-01-02T15:04:05"),
		"merchant_legal_name":    payout.MerchantId,
		"merchant_address":       payout.Destination.Address,
		"merchant_eu_vat_number": "REPLACE_ME!!!",
		"merchant_bank_details":  payout.Destination.Details,
		"period_from":            payout.PeriodFrom.Format("2006-01-02T15:04:05"),
		"period_to":              payout.PeriodTo.Format("2006-01-02T15:04:05"),
		//"transactions_for_period": payout.Summary.Orders.Count,
		"agreement_number": "REPLACE_ME!!!",
		"total_fees":       "REPLACE_ME!!!",
		"balance":          "REPLACE_ME!!!",
	}

	return result, nil
}

func (h *Payout) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	billingService := billingGrpc.NewBillingService(billingProto.ServiceName, h.service.Client())

	req := &billingGrpc.PayoutDocumentPdfUploadedRequest{
		Id:            id,
		Filename:      fileName,
		RetentionTime: int32(retentionTime),
	}

	if _, err := billingService.PayoutDocumentPdfUploaded(ctx, req); err != nil {
		return err
	}

	return nil
}
