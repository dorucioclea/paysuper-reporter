package builder

import (
	"context"
	"errors"
	"fmt"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
)

const (
	errorRequestParameterIsRequired = `parameter "%s"" is required`
	errorRequestParameterIsEmpty    = `parameter "%s"" can't be empty`
)

var (
	agreementRequestRequiredFields = []string{
		pkg.RequestParameterAgreementNumber,
		pkg.RequestParameterAgreementLegalName,
		pkg.RequestParameterAgreementAddress,
		pkg.RequestParameterAgreementRegistrationNumber,
		pkg.RequestParameterAgreementPayoutCost,
		pkg.RequestParameterAgreementMinimalPayoutLimit,
		pkg.RequestParameterAgreementPayoutCurrency,
		pkg.RequestParameterAgreementPSRate,
		pkg.RequestParameterAgreementHomeRegion,
		pkg.RequestParameterAgreementMerchantAuthorizedName,
		pkg.RequestParameterAgreementMerchantAuthorizedPosition,
		pkg.RequestParameterAgreementProjectsLink,
	}
)

type Agreement struct {
	*Handler
	billingService grpc.BillingService
}

func newAgreementHandler(h *Handler) BuildInterface {
	return &Agreement{
		Handler:        h,
		billingService: grpc.NewBillingService(billingPkg.ServiceName, h.service.Client()),
	}
}

func (h *Agreement) Validate() error {
	params, err := h.GetParams()

	if err != nil {
		return err
	}

	for _, v := range agreementRequestRequiredFields {
		val, ok := params[v]

		if !ok {
			return fmt.Errorf(errorRequestParameterIsRequired, v)
		}

		if valTyped, ok := val.(string); ok && valTyped == "" {
			return fmt.Errorf(errorRequestParameterIsEmpty, v)
		} else if valTyped, ok := val.(float64); ok && valTyped == 0 {
			return fmt.Errorf(errorRequestParameterIsEmpty, v)
		}
	}

	return nil
}

func (h *Agreement) Build() (interface{}, error) {
	return h.GetParams()
}

func (h *Agreement) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	req := &grpc.SetMerchantS3AgreementRequest{
		MerchantId:      h.report.MerchantId,
		S3AgreementName: fileName,
	}
	rsp, err := h.billingService.SetMerchantS3Agreement(ctx, req)

	if err != nil {
		return err
	}

	if rsp.Status != billingPkg.ResponseStatusOk {
		return errors.New(rsp.Message.Message)
	}

	return nil
}
