package builder

import (
	"context"
	"encoding/json"
	errs "errors"
	"github.com/micro/go-micro"
	"github.com/paysuper/paysuper-reporter/internal/repository"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
)

var (
	builders = map[string]func(*Handler) BuildInterface{
		pkg.ReportTypeVat:                 newVatHandler,
		pkg.ReportTypeVatTransactions:     newVatTransactionsHandler,
		pkg.ReportTypeRoyalty:             newRoyaltyHandler,
		pkg.ReportTypeRoyaltyTransactions: newRoyaltyTransactionsHandler,
		pkg.ReportTypeTransactions:        newTransactionsHandler,
		pkg.ReportTypePayout:              newPayoutHandler,
	}
)

type BuildInterface interface {
	Validate() error
	Build() (interface{}, error)
	PostProcess(context.Context, string, string, int) error
}

type Handler struct {
	service                micro.Service
	report                 *proto.ReportFile
	royaltyRepository      repository.RoyaltyRepositoryInterface
	vatRepository          repository.VatRepositoryInterface
	transactionsRepository repository.TransactionsRepositoryInterface
	payoutRepository       repository.PayoutRepositoryInterface
}

type DefaultHandler struct {
	*Handler
}

func NewBuilder(
	service micro.Service,
	report *proto.ReportFile,
	royaltyRepository repository.RoyaltyRepositoryInterface,
	vatRepository repository.VatRepositoryInterface,
	transactionsRepository repository.TransactionsRepositoryInterface,
	payoutRepository repository.PayoutRepositoryInterface,
) *Handler {
	return &Handler{
		service:                service,
		report:                 report,
		royaltyRepository:      royaltyRepository,
		vatRepository:          vatRepository,
		transactionsRepository: transactionsRepository,
		payoutRepository:       payoutRepository,
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
