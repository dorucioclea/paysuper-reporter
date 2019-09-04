package builder

import (
	errs "errors"
	"github.com/paysuper/paysuper-reporter/internal/repository"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
)

var (
	builders = map[string]func(*Handler) BuildInterface{
		pkg.ReportTypeTransactions: newTransactionsHandler,
		pkg.ReportTypeVat:          newVatHandler,
		pkg.ReportTypeRoyalty:      newRoyaltyHandler,
	}
)

type BuildInterface interface {
	Validate() error
	Build() (interface{}, error)
}

type Handler struct {
	report                  *proto.MgoReportFile
	reportFileRepository    repository.ReportFileRepositoryInterface
	royaltyReportRepository repository.RoyaltyReportRepositoryInterface
	vatReportRepository     repository.VatReportRepositoryInterface
}

type DefaultHandler struct {
	*Handler
}

func NewBuilder(
	report *proto.MgoReportFile,
	reportFileRepository repository.ReportFileRepositoryInterface,
	royaltyReportRepository repository.RoyaltyReportRepositoryInterface,
	vatReportRepository repository.VatReportRepositoryInterface,
) *Handler {
	return &Handler{
		report:                  report,
		reportFileRepository:    reportFileRepository,
		royaltyReportRepository: royaltyReportRepository,
		vatReportRepository:     vatReportRepository,
	}
}

func (h *Handler) GetBuilder() (BuildInterface, error) {
	handler, ok := builders[h.report.ReportType]

	if !ok {
		return nil, errs.New(errors.ErrorHandlerNotFound.Message)
	}

	return handler(h), nil
}
