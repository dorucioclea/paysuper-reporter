package internal

import (
	"context"
	"encoding/json"
	errs "errors"
	natsMocks "github.com/ProtocolONE/nats/pkg/mocks"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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

func (suite *ReportTestSuite) TestReport_CreateFile_Error_Publish() {
	res := &proto.CreateFileResponse{}
	params, _ := json.Marshal(map[string]interface{}{pkg.ParamsFieldCountry: "RU"})
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
		Params:     params,
	}

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
	params, _ := json.Marshal(map[string]interface{}{pkg.ParamsFieldCountry: "RU"})
	report := &proto.ReportFile{
		ReportType: pkg.ReportTypeVat,
		MerchantId: bson.NewObjectId().Hex(),
		Params:     params,
	}

	broker := &natsMocks.NatsManagerInterface{}
	broker.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	suite.service.messageBroker = broker

	err := suite.service.CreateFile(context.TODO(), report, res)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), pkg.ResponseStatusOk, res.Status)
	assert.NotEmpty(suite.T(), res.FileId)
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
