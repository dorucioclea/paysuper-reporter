package internal

import (
	"context"
	errs "errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	"github.com/paysuper/paysuper-reporter/internal/builder"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"sort"
	"time"
)

var (
	reportTypes = []string{
		pkg.ReportTypeVat,
		pkg.ReportTypeVatTransactions,
		pkg.ReportTypeRoyalty,
		pkg.ReportTypeRoyaltyTransactions,
		pkg.ReportTypeTransactions,
	}

	reportFileContentTypes = map[string]string{
		pkg.OutputExtensionXlsx: pkg.OutputContentTypeXlsx,
		pkg.OutputExtensionCsv:  pkg.OutputContentTypeCsv,
		pkg.OutputExtensionPdf:  pkg.OutputContentTypePdf,
	}

	reportFileRecipes = map[string]string{
		pkg.OutputExtensionXlsx: pkg.RecipeXlsx,
		pkg.OutputExtensionCsv:  pkg.RecipeCsv,
		pkg.OutputExtensionPdf:  pkg.RecipePdf,
	}
)

type ReportFileTemplate struct {
	TemplateId string
	Table      string
	Fields     []string
	Match      string
	Group      string
}

func (app *Application) CreateFile(ctx context.Context, file *proto.ReportFile, res *proto.CreateFileResponse) error {
	var err error

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
	mgoReport, err := file.GetBSON()

	if err != nil {
		zap.L().Error(errors.ErrorConvertBson.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorConvertBson

		return nil
	}

	report := mgoReport.(*proto.MgoReportFile)
	report.ExpireAt = time.Now().Add(time.Duration(app.cfg.DocumentRetentionTime) * time.Second)

	h := builder.NewBuilder(
		report,
		app.reportFileRepository,
		app.royaltyRepository,
		app.vatRepository,
		app.transactionsRepository,
	)
	bldr, err := h.GetBuilder()

	if err != nil {
		zap.L().Error(errors.ErrorHandlerNotFound.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorHandlerNotFound

		return nil
	}

	if err = bldr.Validate(); err != nil {
		zap.L().Error(errors.ErrorHandlerValidation.Message, zap.Error(err), zap.Any("file", mgoReport))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorHandlerValidation

		return nil
	}

	if err := app.reportFileRepository.Insert(report); err != nil {
		zap.L().Error(errors.ErrorUnableToCreate.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorUnableToCreate
		return nil
	}

	if err := app.messageBroker.Publish(pkg.SubjectRequestReportFileCreate, mgoReport, false); err != nil {
		zap.L().Error(errors.ErrorMessageBrokerFailed.Message, zap.Error(err), zap.Any("file", report))
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorMessageBrokerFailed
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.FileId = file.Id

	return nil
}

func (app *Application) LoadFile(ctx context.Context, req *proto.LoadFileRequest, res *proto.LoadFileResponse) error {
	file, err := app.reportFileRepository.GetById(req.Id)

	if err != nil {
		zap.L().Error(errors.ErrorNotFound.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	fileName := fmt.Sprintf(pkg.FileMask, file.Id.Hex(), file.FileType)
	filePath := os.TempDir() + string(os.PathSeparator) + fileName

	if _, err = app.s3.Download(ctx, filePath, &awsWrapper.DownloadInput{FileName: fileName}); err != nil {
		zap.L().Error(errors.ErrorAwsFileNotFound.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorAwsFileNotFound
		return nil
	}

	f, err := os.Open(filePath)
	defer f.Close()

	if err != nil {
		zap.L().Error(errors.ErrorOpenTemporaryFile.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorOpenTemporaryFile
		return nil
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		zap.L().Error(errors.ErrorReadTemporaryFile.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorReadTemporaryFile
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.File = &proto.File{File: b}
	res.ContentType = reportFileContentTypes[file.FileType]
	res.FileType = file.FileType

	_ = os.Remove(filePath)

	return nil
}

func (app *Application) getTemplate(file *proto.ReportFile) (string, error) {
	if file.Template != "" {
		return file.Template, nil
	}

	switch file.ReportType {
	case pkg.ReportTypeRoyalty:
		return app.cfg.DG.RoyaltyTemplate, nil
	case pkg.ReportTypeRoyaltyTransactions:
		return app.cfg.DG.RoyaltyTransactionsTemplate, nil
	case pkg.ReportTypeVat:
		return app.cfg.DG.VatTemplate, nil
	case pkg.ReportTypeVatTransactions:
		return app.cfg.DG.VatTransactionsTemplate, nil
	case pkg.ReportTypeTransactions:
		return app.cfg.DG.TransactionsTemplate, nil
	}

	return file.Template, errs.New(errors.ErrorTemplateNotFound.Message)
}
