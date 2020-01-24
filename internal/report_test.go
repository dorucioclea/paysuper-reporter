package internal

import (
	"context"
	"encoding/json"
	errs "errors"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	rabbitmqMock "gopkg.in/ProtocolONE/rabbitmq.v1/pkg/mocks"
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
	res := &reporterpb.CreateFileResponse{}
	err := suite.service.CreateFile(context.TODO(), &reporterpb.ReportFile{FileType: reporterpb.OutputExtensionPdf}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusBadData, res.Status)
	assert.Equal(suite.T(), errors.ErrorReportTypeNotFound, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_FileType() {
	res := &reporterpb.CreateFileResponse{}
	err := suite.service.CreateFile(context.TODO(), &reporterpb.ReportFile{}, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusBadData, res.Status)
	assert.Equal(suite.T(), errors.ErrorFileType, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_BuilderValidate() {
	res := &reporterpb.CreateFileResponse{}
	report := &reporterpb.ReportFile{
		ReportType: reporterpb.ReportTypeVat,
		FileType:   reporterpb.OutputExtensionPdf,
		MerchantId: "ffffffffffffffffffffffff",
	}
	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusBadData, res.Status)
	assert.Equal(suite.T(), errors.ErrorHandlerValidation, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Error_Publish() {
	res := &reporterpb.CreateFileResponse{}
	params, _ := json.Marshal(map[string]interface{}{reporterpb.ParamsFieldCountry: "RU"})
	report := &reporterpb.ReportFile{
		FileType:   reporterpb.OutputExtensionPdf,
		ReportType: reporterpb.ReportTypeVat,
		MerchantId: "ffffffffffffffffffffffff",
		Params:     params,
	}

	broker := &rabbitmqMock.BrokerInterface{}
	broker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(errs.New("error"))
	suite.service.generateReportBroker = broker

	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusSystemError, res.Status)
	assert.Equal(suite.T(), errors.ErrorMessageBrokerFailed, res.Message)
	assert.Equal(suite.T(), "", res.FileId)
}

func (suite *ReportTestSuite) TestReport_CreateFile_Ok() {
	res := &reporterpb.CreateFileResponse{}
	params, _ := json.Marshal(map[string]interface{}{reporterpb.ParamsFieldCountry: "RU"})
	report := &reporterpb.ReportFile{
		ReportType: reporterpb.ReportTypeVat,
		FileType:   reporterpb.OutputExtensionPdf,
		MerchantId: "ffffffffffffffffffffffff",
		Params:     params,
	}

	broker := &rabbitmqMock.BrokerInterface{}
	broker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.service.generateReportBroker = broker

	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusOk, res.Status)
	assert.NotEmpty(suite.T(), res.FileId)
}

func (suite *ReportTestSuite) TestReport_getTemplate_NotEmptyTemplate() {
	report := &reporterpb.ReportFile{Template: "test"}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultRoyaltyTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		RoyaltyTemplate: "royalty",
	}
	report := &reporterpb.ReportFile{ReportType: reporterpb.ReportTypeRoyalty}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "royalty", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultRoyaltyTransactionsTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		RoyaltyTransactionsTemplate: "royalty_transactions",
	}
	report := &reporterpb.ReportFile{ReportType: reporterpb.ReportTypeRoyaltyTransactions}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "royalty_transactions", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultVatTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		VatTemplate: "vat",
	}
	report := &reporterpb.ReportFile{ReportType: reporterpb.ReportTypeVat}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "vat", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultVatTransactionsTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		VatTransactionsTemplate: "vat_transactions",
	}
	report := &reporterpb.ReportFile{ReportType: reporterpb.ReportTypeVatTransactions}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "vat_transactions", name)
}

func (suite *ReportTestSuite) TestReport_getTemplate_DefaultTransactionsTemplate() {
	suite.service.cfg.DG = config.DocumentGeneratorConfig{
		TransactionsTemplate: "transactions",
	}
	report := &reporterpb.ReportFile{ReportType: reporterpb.ReportTypeTransactions}
	name, err := suite.service.getTemplate(report)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "transactions", name)
}
