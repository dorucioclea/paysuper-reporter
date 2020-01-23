package internal

import (
	"context"
	errs "errors"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/internal/builder"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"sort"
)

var (
	reportTypes = []string{
		reporterpb.ReportTypeVat,
		reporterpb.ReportTypeVatTransactions,
		reporterpb.ReportTypeRoyalty,
		reporterpb.ReportTypeRoyaltyTransactions,
		reporterpb.ReportTypeTransactions,
		reporterpb.ReportTypeAgreement,
	}

	reportFileContentTypes = map[string]string{
		reporterpb.OutputExtensionXlsx: pkg.OutputContentTypeXlsx,
		reporterpb.OutputExtensionCsv:  pkg.OutputContentTypeCsv,
		reporterpb.OutputExtensionPdf:  pkg.OutputContentTypePdf,
	}

	reportFileRecipes = map[string]string{
		reporterpb.OutputExtensionXlsx: pkg.RecipeXlsx,
		reporterpb.OutputExtensionCsv:  pkg.RecipeCsv,
		reporterpb.OutputExtensionPdf:  pkg.RecipePdf,
	}
)

type ReportFileTemplate struct {
	TemplateId string
	Table      string
	Fields     []string
	Match      string
	Group      string
}

func (app *Application) CreateFile(ctx context.Context, file *reporterpb.ReportFile, res *reporterpb.CreateFileResponse) error {
	var err error

	if _, ok := reportFileContentTypes[file.FileType]; !ok {
		zap.L().Error(errors.ErrorFileType.Message, zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorFileType

		return nil
	}

	sort.Strings(reportTypes)

	if file.ReportType == "" || sort.SearchStrings(reportTypes, file.ReportType) == len(reportTypes) {
		zap.L().Error(errors.ErrorReportTypeNotFound.Message, zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorReportTypeNotFound

		return nil
	}

	if file.Template, err = app.getTemplate(file); err != nil {
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorTemplateNotFound

		return nil
	}

	file.Id = bson.NewObjectId().Hex()

	h := builder.NewBuilder(
		app.service,
		file,
		app.billing,
	)
	bldr, err := h.GetBuilder()

	if err != nil {
		zap.L().Error(errors.ErrorHandlerNotFound.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorHandlerNotFound

		return nil
	}

	if err = bldr.Validate(); err != nil {
		zap.L().Error(errors.ErrorHandlerValidation.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorHandlerValidation

		return nil
	}

	amqpHeaders := amqp.Table{
		"x-retry-count": int32(0),
	}
	err = app.generateReportBroker.Publish(pkg.BrokerGenerateReportTopicName, file, amqpHeaders)

	if err != nil {
		zap.L().Error(
			errors.ErrorMessageBrokerFailed.Message,
			zap.Error(err),
			zap.Any("file", file),
		)
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorMessageBrokerFailed
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.FileId = file.Id

	return nil
}

func (app *Application) getTemplate(file *reporterpb.ReportFile) (string, error) {
	if file.Template != "" {
		return file.Template, nil
	}

	switch file.ReportType {
	case reporterpb.ReportTypeRoyalty:
		return app.cfg.DG.RoyaltyTemplate, nil
	case reporterpb.ReportTypeRoyaltyTransactions:
		return app.cfg.DG.RoyaltyTransactionsTemplate, nil
	case reporterpb.ReportTypeVat:
		return app.cfg.DG.VatTemplate, nil
	case reporterpb.ReportTypeVatTransactions:
		return app.cfg.DG.VatTransactionsTemplate, nil
	case reporterpb.ReportTypeTransactions:
		return app.cfg.DG.TransactionsTemplate, nil
	case reporterpb.ReportTypeAgreement:
		return app.cfg.DG.AgreementTemplate, nil
	case reporterpb.ReportTypePayout:
		return app.cfg.DG.PayoutTemplate, nil
	}

	return file.Template, errs.New(errors.ErrorTemplateNotFound.Message)
}
