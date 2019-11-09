package builder

import (
	"context"
	"errors"
	"fmt"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billingGrpc "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
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

type Agreement DefaultHandler

type TariffPrintable struct {
	MinAmount    string `json:"min_amount"`
	MaxAmount    string `json:"max_amount"`
	MethodName   string `json:"method_name"`
	PsPercentFee string `json:"ps_percent_fee"`
	PsFixedFee   string `json:"ps_fixed_fee"`
	PayerRegion  string `json:"payer_region"`
}

func newAgreementHandler(h *Handler) BuildInterface {
	return &Agreement{Handler: h}
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

func (h *Agreement) PostProcess(ctx context.Context, id string, fileName string, retentionTime int, content []byte) error {
	req := &billingGrpc.SetMerchantS3AgreementRequest{
		MerchantId:      h.report.MerchantId,
		S3AgreementName: fileName,
	}
	rsp, err := h.billing.SetMerchantS3Agreement(ctx, req)

	if err != nil {
		return err
	}

	if rsp.Status != billingPkg.ResponseStatusOk {
		return errors.New(rsp.Message.Message)
	}

	return nil
}
