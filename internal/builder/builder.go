package builder

import (
	"context"
	"encoding/json"
	errs "errors"
	"github.com/micro/go-micro"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
)

var (
	builders = map[string]func(*Handler) BuildInterface{
		reporterpb.ReportTypeVat:                 newVatHandler,
		reporterpb.ReportTypeVatTransactions:     newVatTransactionsHandler,
		reporterpb.ReportTypeRoyalty:             newRoyaltyHandler,
		reporterpb.ReportTypeRoyaltyTransactions: newRoyaltyTransactionsHandler,
		reporterpb.ReportTypeTransactions:        newTransactionsHandler,
		reporterpb.ReportTypePayout:              newPayoutHandler,
		reporterpb.ReportTypeAgreement:           newAgreementHandler,
	}
)

type BuildInterface interface {
	Validate() error
	Build() (interface{}, error)
	PostProcess(context.Context, string, string, int64, []byte) error
}

type Handler struct {
	service micro.Service
	report  *reporterpb.ReportFile
	billing billingpb.BillingService
}

type DefaultHandler struct {
	*Handler
}

func NewBuilder(
	service micro.Service,
	report *reporterpb.ReportFile,
	billing billingpb.BillingService,
) *Handler {
	return &Handler{
		service: service,
		report:  report,
		billing: billing,
	}
}

func (h *Handler) GetBuilder() (BuildInterface, error) {
	handler, ok := builders[h.report.ReportType]

	if !ok {
		return nil, errs.New(errors.ErrorHandlerNotFound.Message)
	}

	return handler(h), nil
}

func (h *Handler) GetParams() (map[string]interface{}, error) {
	var params map[string]interface{}

	if h.report.Params == nil {
		return params, nil
	}

	if err := json.Unmarshal(h.report.Params, &params); err != nil {
		return nil, err
	}

	return params, nil
}
