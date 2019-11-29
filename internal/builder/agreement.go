package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/micro/go-micro/client"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
	"time"
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

type TariffPrintable struct {
	MinAmount    string `json:"min_amount"`
	MaxAmount    string `json:"max_amount"`
	MethodName   string `json:"method_name"`
	PsPercentFee string `json:"ps_percent_fee"`
	PsFixedFee   string `json:"ps_fixed_fee"`
	PayerRegion  string `json:"payer_region"`
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
	params, err := h.GetParams()

	if err != nil {
		return nil, err
	}

	var tariffsPrintable []*TariffPrintable
	tariffs := params[pkg.RequestParameterAgreementPSRate].([]interface{})

	for _, v := range tariffs {
		vTyped := v.(map[string]interface{})
		maxAmount := vTyped["max_amount"].(float64)
		tariff := &TariffPrintable{
			MinAmount:    fmt.Sprintf("%.2f", vTyped["min_amount"]),
			MaxAmount:    fmt.Sprintf("%.2f", maxAmount),
			MethodName:   vTyped["method_name"].(string),
			PsPercentFee: fmt.Sprintf("%.2f", vTyped["ps_percent_fee"]),
			PsFixedFee:   fmt.Sprintf("%.2f", vTyped["ps_fixed_fee"]),
			PayerRegion:  billingPkg.HomeRegions[vTyped["payer_region"].(string)],
		}

		if maxAmount == 99999999 {
			tariff.MaxAmount = "..."
		}

		tariffsPrintable = append(tariffsPrintable, tariff)
	}

	params[pkg.RequestParameterAgreementPSRate] = tariffsPrintable

	return params, nil
}

func (h *Agreement) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	req := &grpc.SetMerchantS3AgreementRequest{
		MerchantId:      h.report.MerchantId,
		S3AgreementName: fileName,
	}

	ctx, _ = context.WithTimeout(context.Background(), time.Minute*2)
	opts := []client.CallOption{
		client.WithRequestTimeout(time.Minute * 2),
	}
	rsp, err := h.billingService.SetMerchantS3Agreement(ctx, req, opts...)

	if err != nil {
		return err
	}

	if rsp.Status != billingPkg.ResponseStatusOk {
		return errors.New(rsp.Message.Message)
	}

	return nil
}
