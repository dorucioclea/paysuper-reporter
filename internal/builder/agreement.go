package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"time"
)

const (
	errorRequestParameterIsRequired = `parameter "%s"" is required`
	errorRequestParameterIsEmpty    = `parameter "%s"" can't be empty`

	paymentAmountCurrency = "USD"
)

var (
	agreementRequestRequiredFields = []string{
		reporterpb.RequestParameterAgreementNumber,
		reporterpb.RequestParameterAgreementLegalName,
		reporterpb.RequestParameterAgreementAddress,
		reporterpb.RequestParameterAgreementRegistrationNumber,
		reporterpb.RequestParameterAgreementPayoutCost,
		reporterpb.RequestParameterAgreementMinimalPayoutLimit,
		reporterpb.RequestParameterAgreementPayoutCurrency,
		reporterpb.RequestParameterAgreementPSRate,
		reporterpb.RequestParameterAgreementHomeRegion,
		reporterpb.RequestParameterAgreementMerchantAuthorizedName,
		reporterpb.RequestParameterAgreementMerchantAuthorizedPosition,
		reporterpb.RequestParameterAgreementOperatingCompanyLegalName,
		reporterpb.RequestParameterAgreementOperatingCompanyAddress,
		reporterpb.RequestParameterAgreementOperatingCompanyRegistrationNumber,
		reporterpb.RequestParameterAgreementOperatingCompanyAuthorizedName,
		reporterpb.RequestParameterAgreementOperatingCompanyAuthorizedPosition,
	}
)

type AgreementInterface interface {
	GetAgreementName(fileType string) (string, error)
}

type Agreement DefaultHandler

type TariffPrintable struct {
	Region                string `json:"payer_region"`
	MethodName            string `json:"method_name"`
	PaymentAmountMin      string `json:"payment_amount_min"`
	PaymentAmountMax      string `json:"payment_amount_max"`
	PaymentAmountCurrency string `json:"payment_amount_currency"`
	PsPercentFee          string `json:"ps_percent_fee"`
	PsFixedFee            string `json:"ps_fixed_fee"`
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
	tariffs := params[reporterpb.RequestParameterAgreementPSRate].([]interface{})

	for _, v := range tariffs {
		vTyped := v.(map[string]interface{})
		tariff := &TariffPrintable{
			Region:                vTyped["payer_region"].(string),
			MethodName:            vTyped["method_name"].(string),
			PaymentAmountMin:      fmt.Sprintf("%.2f", vTyped["min_amount"]),
			PaymentAmountMax:      fmt.Sprintf("%.2f", vTyped["max_amount"]),
			PaymentAmountCurrency: paymentAmountCurrency,
			PsPercentFee:          fmt.Sprintf("%.2f", vTyped["ps_percent_fee"].(float64)*100),
			PsFixedFee:            fmt.Sprintf("%.2f", vTyped["ps_fixed_fee"]),
		}

		tariffsPrintable = append(tariffsPrintable, tariff)
	}

	params[reporterpb.RequestParameterAgreementPSRate] = tariffsPrintable

	return params, nil
}

func (h *Agreement) PostProcess(
	ctx context.Context,
	_ string,
	fileName string,
	_ int64,
	_ []byte,
) error {
	req := &billingpb.SetMerchantS3AgreementRequest{
		MerchantId:      h.report.MerchantId,
		S3AgreementName: fileName,
	}

	opts := []client.CallOption{
		client.WithRequestTimeout(time.Minute * 2),
	}
	rsp, err := h.billing.SetMerchantS3Agreement(ctx, req, opts...)

	if err != nil {
		return err
	}

	if rsp.Status != billingpb.ResponseStatusOk {
		return errors.New(rsp.Message.Message)
	}

	return nil
}

func (h *Agreement) GetAgreementName(fileType string) (string, error) {
	params, err := h.GetParams()

	if err != nil {
		return "", err
	}

	name := fmt.Sprintf(reporterpb.FileMaskAgreement, params[reporterpb.RequestParameterAgreementLegalName], params[reporterpb.RequestParameterAgreementNumber], fileType)
	return name, nil
}
