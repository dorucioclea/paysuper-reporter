package internal

import (
	"context"
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
)

var (
	reportTypes = []string{
		pkg.ReportTypeVat,
		pkg.ReportTypeVatTransactions,
		pkg.ReportTypeRoyalty,
		pkg.ReportTypeRoyaltyTransactions,
		pkg.ReportTypeTransactions,
	}

	reportFileTypes = map[string]string{
		pkg.OutputXlsxExtension: pkg.OutputXlsxContentType,
		pkg.OutputCsvExtension:  pkg.OutputCsvContentType,
		pkg.OutputPdfExtension:  pkg.OutputPdfContentType,
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
	fmt.Println(sort.SearchStrings(reportTypes, file.ReportType))
	fmt.Println(len(reportTypes))
	fmt.Println(reportTypes)
	fmt.Println(file.ReportType)
	if sort.SearchStrings(reportTypes, file.ReportType) == len(reportTypes) {
		zap.L().Error(errors.ErrorReportTypeNotFound.Message, zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorReportTypeNotFound

		return nil
	}

	if _, ok := reportFileTypes[file.FileType]; !ok {
		zap.L().Error(errors.ErrorFileTypeNotFound.Message, zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorFileTypeNotFound

		return nil
	}

	if file.Template == "" {
		switch file.ReportType {
		case pkg.ReportTypeRoyalty:
			file.Template = app.cfg.DG.RoyaltyTemplate
		case pkg.ReportTypeRoyaltyTransactions:
			file.Template = app.cfg.DG.RoyaltyTransactionsTemplate
		case pkg.ReportTypeVat:
			file.Template = app.cfg.DG.VatTemplate
		case pkg.ReportTypeVatTransactions:
			file.Template = app.cfg.DG.VatTransactionsTemplate
		case pkg.ReportTypeTransactions:
			file.Template = app.cfg.DG.TransactionsTemplate
		}
	}

	if file.Id == "" {
		file.Id = bson.NewObjectId().Hex()
	}

	mgoReport, err := file.GetBSON()

	if err != nil {
		zap.L().Error(errors.ErrorConvertBson.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorConvertBson

		return nil
	}

	h := builder.NewBuilder(
		mgoReport.(*proto.MgoReportFile),
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

	if err := app.reportFileRepository.Insert(file); err != nil {
		zap.L().Error(errors.ErrorUnableToCreate.Message, zap.Error(err), zap.Any("file", file))
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorUnableToCreate
		return nil
	}

	if err := app.messageBroker.Publish(pkg.SubjectRequestReportFileCreate, mgoReport, false); err != nil {
		zap.L().Error(errors.ErrorMessageBrokerFailed.Message, zap.Error(err), zap.Any("file", mgoReport))
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

	filePath := os.TempDir() + string(os.PathSeparator) + file.Id

	if _, err = app.s3.Download(ctx, filePath, &awsWrapper.DownloadInput{FileName: file.Id}); err != nil {
		zap.L().Error(errors.ErrorNotFound.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	f, err := os.Open(filePath)
	defer f.Close()

	if err != nil {
		zap.L().Error(errors.ErrorNotFound.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		zap.L().Error(errors.ErrorNotFound.Message, zap.Error(err), zap.Any("data", req))
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.File.File = b

	return nil
}
