package internal

import (
	"context"
	"encoding/json"
	errs "errors"
	"fmt"
	natsMocks "github.com/ProtocolONE/nats/pkg/mocks"
	"github.com/globalsign/mgo/bson"
	awsMocks "github.com/paysuper/paysuper-aws-manager/pkg/mocks"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

type ReportTestSuite struct {
	suite.Suite
	service *Application
}

func Test_Report(t *testing.T) {
	suite.Run(t, new(ReportTestSuite))
}

func (suite *ReportTestSuite) SetupTest() {
	suite.service = &Application{cfg: &config.Config{}}
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_ReportType() {
	res := &proto.CreateFileResponse{}
	err := suite.service.CreateFile(context.TODO(), &proto.ReportFile{}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusBadData, res.Status)
	assert.Equal(suite.T(), errors.ErrorReportTypeNotFound, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_UnableBsonParams() {
	res := &proto.CreateFileResponse{}
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
		Params:     []byte{},
	}
	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusBadData, res.Status)
	assert.Equal(suite.T(), errors.ErrorConvertBson, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_BuilderValidate() {
	res := &proto.CreateFileResponse{}
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
	}
	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusBadData, res.Status)
	assert.Equal(suite.T(), errors.ErrorHandlerValidation, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_InsertReportFile() {
	res := &proto.CreateFileResponse{}
	params, _ := json.Marshal(map[string]interface{}{pkg.ParamsFieldId: 1})
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
		Params:     params,
	}

	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("Insert", mock.Anything).Return(errs.New("error"))
	suite.service.reportFileRepository = repo

	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusSystemError, res.Status)
	assert.Equal(suite.T(), errors.ErrorUnableToCreate, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_Publish() {
	res := &proto.CreateFileResponse{}
	params, _ := json.Marshal(map[string]interface{}{pkg.ParamsFieldId: 1})
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
		Params:     params,
	}

	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("Insert", mock.Anything).Return(nil)
	suite.service.reportFileRepository = repo

	broker := &natsMocks.NatsManagerInterface{}
	broker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errs.New("error"))
	suite.service.messageBroker = broker

	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusSystemError, res.Status)
	assert.Equal(suite.T(), errors.ErrorMessageBrokerFailed, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Ok() {
	res := &proto.CreateFileResponse{}
	params, _ := json.Marshal(map[string]interface{}{pkg.ParamsFieldId: 1})
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
		Params:     params,
	}

	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("Insert", mock.Anything).Return(nil)
	suite.service.reportFileRepository = repo

	broker := &natsMocks.NatsManagerInterface{}
	broker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.service.messageBroker = broker

	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusOk, res.Status)
	assert.NotEmpty(suite.T(), res.FileId)
}

func (suite *ReportTestSuite) TestReport_LoadFile_Error_FileNotFound() {
	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("GetById", mock.Anything).Return(nil, errs.New("not found"))
	suite.service.reportFileRepository = repo

	res := &proto.LoadFileResponse{}
	err := suite.service.LoadFile(context.TODO(), &proto.LoadFileRequest{}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusNotFound, res.Status)
	assert.Equal(suite.T(), errors.ErrorNotFound, res.Message)
	assert.Empty(suite.T(), res.ContentType)
	assert.Nil(suite.T(), res.File)
}

func (suite *ReportTestSuite) TestReport_LoadFile_Error_S3Download() {
	file := &proto.MgoReportFile{
		Id:       bson.NewObjectId(),
		FileType: pkg.OutputExtensionPdf,
	}
	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("GetById", mock.Anything).Return(file, nil)
	suite.service.reportFileRepository = repo

	s3 := &awsMocks.AwsManagerInterface{}
	s3.On("Download", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(0), errs.New("not found"))
	suite.service.s3 = s3

	res := &proto.LoadFileResponse{}
	err := suite.service.LoadFile(context.TODO(), &proto.LoadFileRequest{}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusNotFound, res.Status)
	assert.Equal(suite.T(), errors.ErrorAwsFileNotFound, res.Message)
	assert.Empty(suite.T(), res.ContentType)
	assert.Nil(suite.T(), res.File)
}

func (suite *ReportTestSuite) TestReport_LoadFile_Error_OpenFile() {
	file := &proto.MgoReportFile{
		Id:       bson.NewObjectId(),
		FileType: pkg.OutputExtensionPdf,
	}
	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("GetById", mock.Anything).Return(file, nil)
	suite.service.reportFileRepository = repo

	s3 := &awsMocks.AwsManagerInterface{}
	s3.On("Download", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	suite.service.s3 = s3

	res := &proto.LoadFileResponse{}
	err := suite.service.LoadFile(context.TODO(), &proto.LoadFileRequest{}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusNotFound, res.Status)
	assert.Equal(suite.T(), errors.ErrorOpenTemporaryFile, res.Message)
	assert.Empty(suite.T(), res.ContentType)
	assert.Nil(suite.T(), res.File)
}

func (suite *ReportTestSuite) TestReport_LoadFile_Ok() {
	file := &proto.MgoReportFile{
		Id:       bson.NewObjectId(),
		FileType: pkg.OutputExtensionPdf,
	}
	repo := &mocks.ReportFileRepositoryInterface{}
	repo.On("GetById", mock.Anything).Return(file, nil)
	suite.service.reportFileRepository = repo

	s3 := &awsMocks.AwsManagerInterface{}
	s3.On("Download", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	suite.service.s3 = s3

	fileName := fmt.Sprintf(pkg.FileMask, file.Id.Hex(), file.FileType)
	filePath := os.TempDir() + string(os.PathSeparator) + fileName
	_ = ioutil.WriteFile(filePath, nil, 0644)

	res := &proto.LoadFileResponse{}
	err := suite.service.LoadFile(context.TODO(), &proto.LoadFileRequest{}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusOk, res.Status)
	assert.Equal(suite.T(), pkg.OutputContentTypePdf, res.ContentType)
	assert.NotEmpty(suite.T(), res.File)
}

func (suite *ReportTestSuite) TestReport_getTemplate_NotEmptyTemplate() {
	report := &proto.ReportFile{Template: "test"}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultRoyaltyTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		RoyaltyTemplate: "royalty",
	}
	report := &proto.ReportFile{ReportType: pkg.ReportTypeRoyalty}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "royalty", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultRoyaltyTransactionsTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		RoyaltyTransactionsTemplate: "royalty_transactions",
	}
	report := &proto.ReportFile{ReportType: pkg.ReportTypeRoyaltyTransactions}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "royalty_transactions", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultVatTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		VatTemplate: "vat",
	}
	report := &proto.ReportFile{ReportType: pkg.ReportTypeVat}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "vat", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultVatTransactionsTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		VatTransactionsTemplate: "vat_transactions",
	}
	report := &proto.ReportFile{ReportType: pkg.ReportTypeVatTransactions}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "vat_transactions", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultTransactionsTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		TransactionsTemplate: "transactions",
	}
	report := &proto.ReportFile{ReportType: pkg.ReportTypeTransactions}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "transactions", name)
}
