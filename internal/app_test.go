package internal

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	awsWrapperMocks "github.com/paysuper/paysuper-aws-manager/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	rabbitmqMock "gopkg.in/ProtocolONE/rabbitmq.v1/pkg/mocks"
	"testing"
)

type ApplicationTestSuite struct {
	suite.Suite
	dummyApp *Application
}

func Test_Application(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (suite *ApplicationTestSuite) SetupTest() {
	awsManagerMock := &awsWrapperMocks.AwsManagerInterface{}
	awsManagerMock.On("Upload", mock2.Anything, mock2.Anything, mock2.Anything).Return(&s3manager.UploadOutput{}, nil)

	centrifugoMock := &mocks.CentrifugoInterface{}
	centrifugoMock.On("Publish", mock2.Anything, mock2.Anything).Return(nil, nil)

	documentGeneratorMock := &mocks.DocumentGeneratorInterface{}
	documentGeneratorMock.On("Render", mock2.Anything).Return([]byte("agreement file content"), nil)

	royaltyRepository := &mocks.RoyaltyRepositoryInterface{}
	royaltyRepository.On("GetById", mock2.Anything).Return(&billing.MgoRoyaltyReport{}, nil)

	vatRepositoryMock := &mocks.VatRepositoryInterface{}
	vatRepositoryMock.On("GetByCountry", mock2.Anything).Return(make([]*billing.MgoVatReport, 1), nil)
	vatRepositoryMock.On("GetById", mock2.Anything).Return(&billing.MgoVatReport{}, nil)

	transactionsRepositoryMock := &mocks.TransactionsRepositoryInterface{}
	transactionsRepositoryMock.On("FindByMerchant", mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything).
		Return(make([]*billing.MgoOrderViewPublic, 1), nil)
	transactionsRepositoryMock.On("GetByRoyalty", mock2.Anything).Return(make([]*billing.MgoOrderViewPublic, 1), nil)
	transactionsRepositoryMock.On("GetByVat", mock2.Anything).Return(make([]*billing.MgoOrderViewPrivate, 1), nil)

	payoutRepositoryMock := &mocks.PayoutRepositoryInterface{}
	payoutRepositoryMock.On("GetById", mock2.Anything).Return(&billing.MgoPayoutDocument{}, nil)

	merchantRepositoryMock := &mocks.MerchantRepositoryInterface{}
	merchantRepositoryMock.On("GetById", mock2.Anything).Return(&billing.MgoMerchant{}, nil)

	brokerMock := &rabbitmqMock.BrokerInterface{}
	brokerMock.On("Publish", mock2.Anything, mock2.Anything, mock2.Anything).Return(nil, nil)
	brokerMock.On("RegisterSubscriber", mock2.Anything, mock2.Anything).Return(nil, nil)
	brokerMock.On("SetExchangeName", mock2.Anything).Return(nil)
	brokerMock.On("Subscribe", mock2.Anything).Return(nil, nil)

	suite.dummyApp = &Application{
		s3:                     awsManagerMock,
		s3Agreement:            awsManagerMock,
		centrifugo:             centrifugoMock,
		documentGenerator:      documentGeneratorMock,
		royaltyRepository:      royaltyRepository,
		vatRepository:          vatRepositoryMock,
		transactionsRepository: transactionsRepositoryMock,
		payoutRepository:       payoutRepositoryMock,
		merchantRepository:     merchantRepositoryMock,
		generateReportBroker:   brokerMock,
		postProcessBroker:      brokerMock,
		cfg: &config.Config{
			S3:               config.S3Config{},
			DG:               config.DocumentGeneratorConfig{},
			CentrifugoConfig: config.CentrifugoConfig{},
		},
	}
}

func (suite *ApplicationTestSuite) TearDownTest() {}

func (suite *ApplicationTestSuite) TestApplication_ExecuteProcess_Agreement_Ok() {
	fileName := ""

	awsUploadMockFn := func(
		ctx context.Context,
		in *awsWrapper.UploadInput,
		opts ...func(*s3manager.Uploader),
	) *s3manager.UploadOutput {
		fileName = in.FileName
		return &s3manager.UploadOutput{}
	}

	awsManagerMock := &awsWrapperMocks.AwsManagerInterface{}
	awsManagerMock.On("Upload", mock2.Anything, mock2.Anything, mock2.Anything).Return(awsUploadMockFn, nil)
	suite.dummyApp.s3Agreement = awsManagerMock

	params := map[string]interface{}{
		reporterPkg.RequestParameterAgreementNumber:             "123456-AA-7890",
		reporterPkg.RequestParameterAgreementLegalName:          "Company Name",
		reporterPkg.RequestParameterAgreementAddress:            "Company address",
		reporterPkg.RequestParameterAgreementRegistrationNumber: "Company registration number",
		reporterPkg.RequestParameterAgreementPayoutCost:         "Payout cost",
		reporterPkg.RequestParameterAgreementMinimalPayoutLimit: "Min payout limit",
		reporterPkg.RequestParameterAgreementPayoutCurrency:     "USD",
		reporterPkg.RequestParameterAgreementPSRate: []map[string]interface{}{
			{
				"min_amount":                1,
				"max_amount":                4.99,
				"method_name":               "VISA",
				"method_percent_fee":        0.09,
				"method_fixed_fee":          9,
				"method_fixed_fee_currency": "USD",
				"ps_percent_fee":            0.05,
				"ps_fixed_fee":              0.3,
				"ps_fixed_fee_currency":     "USD",
				"merchant_home_region":      "latin_america",
				"payer_region":              "asia",
				"mcc_code":                  "5816",
				"is_active":                 true,
			},
			{
				"min_amount":                1,
				"max_amount":                4.99,
				"method_name":               "Mastercard",
				"method_percent_fee":        0.03,
				"method_fixed_fee":          0.2,
				"method_fixed_fee_currency": "USD",
				"ps_percent_fee":            0.05,
				"ps_fixed_fee":              0.3,
				"ps_fixed_fee_currency":     "USD",
				"merchant_home_region":      "latin_america",
				"payer_region":              "asia",
				"mcc_code":                  "5816",
				"is_active":                 true,
			},
		},
		reporterPkg.RequestParameterAgreementHomeRegion:                         "Russia",
		reporterPkg.RequestParameterAgreementMerchantAuthorizedName:             "Authorized Name",
		reporterPkg.RequestParameterAgreementMerchantAuthorizedPosition:         "Authorized Position",
		reporterPkg.RequestParameterAgreementOperatingCompanyLegalName:          "Operating company name",
		reporterPkg.RequestParameterAgreementOperatingCompanyAddress:            "Operating company address",
		reporterPkg.RequestParameterAgreementOperatingCompanyRegistrationNumber: "Operating company registration number",
		reporterPkg.RequestParameterAgreementOperatingCompanyAuthorizedName:     "Operating company signatory name",
		reporterPkg.RequestParameterAgreementOperatingCompanyAuthorizedPosition: "Operating company signatory position",
	}

	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)

	payload := &proto.ReportFile{
		UserId:           "ffffffffffffffffffffffff",
		MerchantId:       "ffffffffffffffffffffffff",
		ReportType:       reporterPkg.ReportTypeAgreement,
		FileType:         reporterPkg.OutputExtensionPdf,
		Params:           b,
		SendNotification: false,
	}
	err = suite.dummyApp.ExecuteProcess(payload, amqp.Delivery{})
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), fileName, "License Agreement_Company Name_#123456-AA-7890.pdf")
}
