package builder

import (
	"encoding/json"
	errs "errors"
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
	}
)

type BuildInterface interface {
	Validate() error
	Build() (interface{}, error)
}

type Handler struct {
	report                  *proto.ReportFile
	royaltyReportRepository repository.RoyaltyRepositoryInterface
	vatReportRepository     repository.VatRepositoryInterface
	transactionsRepository  repository.TransactionsRepositoryInterface
}

type DefaultHandler struct {
	*Handler
}

func NewBuilder(
	report *proto.ReportFile,
	royaltyReportRepository repository.RoyaltyRepositoryInterface,
	vatReportRepository repository.VatRepositoryInterface,
	transactionsRepository repository.TransactionsRepositoryInterface,
) *Handler {
	return &Handler{
		report:                  report,
		royaltyReportRepository: royaltyReportRepository,
		vatReportRepository:     vatReportRepository,
		transactionsRepository:  transactionsRepository,
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
