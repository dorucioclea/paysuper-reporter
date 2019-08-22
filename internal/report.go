package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
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

	reportFileTypes = []string{
		pkg.OutputXslx,
		pkg.OutputCsv,
		pkg.OutputPdf,
	}
)

type ReportFileTemplate struct {
	TemplateId string
	Table      string
	Fields     []string
	Match      string
	Group      string
}

func (app *Application) CreateFile(ctx context.Context, req *proto.CreateFileRequest, res *proto.CreateFileResponse) error {
	template, ok := reportTemplates[req.ReportType]
	if !ok {
		zap.S().Errorf(errors.ErrorTemplateNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorTemplateNotFound

		return nil
	}

	i := sort.SearchStrings(reportFileTypes, req.FileType)
	if i == len(reportFileTypes) {
		zap.S().Errorf(errors.ErrorType.Message, "data", req)
		res.Status = pkg.ResponseStatusBadData
		res.Message = errors.ErrorType

		return nil
	}

	file := &proto.ReportFile{
		Id:         bson.NewObjectId().Hex(),
		MerchantId: req.MerchantId,
		Type:       req.ReportType,
	}
	file.DateFrom, _ = ptypes.TimestampProto(time.Unix(req.PeriodFrom, 0))
	file.DateTo, _ = ptypes.TimestampProto(time.Unix(req.PeriodTo, 0))

	if err := app.reportFileRepository.Insert(file); err != nil {
		zap.S().Errorf(errors.ErrorUnableToCreate.Message, "data", req)
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorUnableToCreate
		return nil
	}

	match, err := json.Marshal(template.Match)
	if err != nil {
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorMarshalMatch
		return err
	}

	group, err := json.Marshal(template.Group)
	if err != nil {
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorMarshalGroup
		return err
	}

	msg := &proto.ReportRequest{
		FileId:       file.Id,
		TemplateId:   template.TemplateId,
		OutputFormat: req.FileType,
		TableName:    template.Table,
		Fields:       template.Fields,
		Match:        match,
		Group:        group,
	}
	if err := app.messageBroker.Publish(pkg.SubjectRequestReportFileCreate, msg, false); err != nil {
		zap.S().Errorf(errors.ErrorMessageBrokerFailed.Message, "data", req)
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorMessageBrokerFailed
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.FileId = file.Id

	return nil
}

func (app *Application) UpdateFile(ctx context.Context, req *proto.UpdateFileRequest, res *proto.ResponseError) error {
	file, err := app.reportFileRepository.GetById(req.Id)

	if err != nil {
		zap.S().Errorf(errors.ErrorNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	err = app.centrifugo.Publish(fmt.Sprintf(app.cfg.CentrifugoConfig.MerchantChannel, file.MerchantId), file)

	if err != nil {
		zap.S().Error(errors.ErrorCentrifugoNotificationFailed, zap.Error(err), zap.Any("report_file", file))
		res.Status = pkg.ResponseStatusSystemError
		res.Message = errors.ErrorCentrifugoNotificationFailed

		return nil
	}

	res.Status = pkg.ResponseStatusOk

	return nil
}

func (app *Application) GetFile(ctx context.Context, req *proto.GetFileRequest, res *proto.GetFileResponse) error {
	file, err := app.reportFileRepository.GetById(req.Id)

	if err != nil {
		zap.S().Errorf(errors.ErrorNotFound.Message, "data", req)
		res.Status = pkg.ResponseStatusNotFound
		res.Message = errors.ErrorNotFound
		return nil
	}

	res.Status = pkg.ResponseStatusOk
	res.File = file

	return nil
}

func (app *Application) LoadFile(ctx context.Context, req *proto.GetFileRequest, res *proto.LoadFileResponse) error {
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
