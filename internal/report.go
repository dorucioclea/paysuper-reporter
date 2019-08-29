package internal

import (
	"context"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/internal/builder"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)

var (
	reportTemplates = map[string]*ReportFileTemplate{
		pkg.ReportTypeVat: {
			TemplateId: pkg.ReportTypeVatTemplate,
			Table:      "",
			Fields:     []string{"payment_system_id", "name", "min_payment_amount", "max_payment_amount"},
			Match: `{
				"is_active": true,
			}`,
			Group: "",
		},
		pkg.ReportTypeTax: {
			TemplateId: pkg.ReportTypeTaxTemplate,
			Table:      "",
			Fields:     []string{},
			Match:      "",
			Group:      "",
		},
		pkg.ReportTypeRoyalty: {
			TemplateId: pkg.ReportTypeRoyaltyTemplate,
			Table:      "",
			Fields:     []string{},
			Match:      "",
			Group:      "",
		},
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
	if file.Template == "" {
		zap.S().Errorf(errors.ErrorTemplateNotFound.Message, "file", file)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorTemplateNotFound

		return nil
	}

	if _, ok := reportTemplates[file.ReportType]; !ok {
		zap.S().Errorf(errors.ErrorReportTypeNotFound.Message, "file", file)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorReportTypeNotFound

		return nil
	}

	if _, ok := reportFileTypes[file.FileType]; !ok {
		zap.S().Errorf(errors.ErrorFileTypeNotFound.Message, "file", file)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorFileTypeNotFound

		return nil
	}

	mgoReport, err := file.GetBSON()

	if err != nil {
		zap.S().Errorf(errors.ErrorConvertBson.Message, "file", file)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorConvertBson

		return nil
	}

	h := builder.NewBuilder(
		mgoReport.(*proto.MgoReportFile),
		app.reportFileRepository,
		app.royaltyReportRepository,
		app.vatReportRepository,
	)
	bldr, err := h.GetBuilder()

	if err != nil {
		zap.S().Errorf(errors.ErrorHandlerNotFound.Message, "file", file)
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorHandlerNotFound

		return nil
	}

	if err = bldr.Validate(); err != nil {
		zap.S().Errorf(errors.ErrorHandlerValidation.Message, "file", mgoReport)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorHandlerValidation

		return nil
	}

	file.Id = bson.NewObjectId().Hex()

	if err := app.reportFileRepository.Insert(file); err != nil {
		zap.S().Errorf(errors.ErrorUnableToCreate.Message, "file", file)
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorUnableToCreate
		return nil
	}

	if err := app.messageBroker.Publish(pkg.SubjectRequestReportFileCreate, mgoReport, false); err != nil {
		zap.S().Errorf(errors.ErrorMessageBrokerFailed.Message, "file", mgoReport)
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorMessageBrokerFailed
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.File = file

	return nil
}

func (app *Application) LoadFile(ctx context.Context, req *proto.LoadFileRequest, res *proto.LoadFileResponse) error {
	file, err := app.reportFileRepository.GetById(req.Id)

	if err != nil {
		zap.S().Errorf(errors.ErrorNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	filePath := os.TempDir() + string(os.PathSeparator) + file.Id
	if err = app.s3.Get(file.Id, filePath, GetObjectOptions{}); err != nil {
		zap.S().Errorf(errors.ErrorNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	f, err := os.Open(filePath)
	defer f.Close()

	if err != nil {
		zap.S().Errorf(errors.ErrorNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		zap.S().Errorf(errors.ErrorNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.File.File = b

	return nil
}
