package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/micro/go-micro"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	billingMocks "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AgreementBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_AgreementBuilder(t *testing.T) {
	suite.Run(t, new(AgreementBuilderTestSuite))
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_NewAgreementBuilder_Ok() {
	builder := newAgreementHandler(&Handler{service: micro.NewService()})
	assert.NotNil(suite.T(), builder)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_Ok() {
	params := map[string]interface{}{
		reporterpb.RequestParameterAgreementNumber:             "123-TEST",
		reporterpb.RequestParameterAgreementLegalName:          "Test",
		reporterpb.RequestParameterAgreementAddress:            "Russia, St.Petersburg, Unit Test 1",
		reporterpb.RequestParameterAgreementRegistrationNumber: "0000000000000000000000001",
		reporterpb.RequestParameterAgreementPayoutCost:         10,
		reporterpb.RequestParameterAgreementMinimalPayoutLimit: 10000,
		reporterpb.RequestParameterAgreementPayoutCurrency:     "USD",
		reporterpb.RequestParameterAgreementPSRate: []*billingpb.MerchantTariffRatesPayment{
			{
				MinAmount:              0,
				MaxAmount:              4.99,
				MethodName:             "VISA",
				MethodPercentFee:       1.8,
				MethodFixedFee:         0.2,
				MethodFixedFeeCurrency: "USD",
				PsPercentFee:           3.0,
				PsFixedFee:             0.3,
				PsFixedFeeCurrency:     "USD",
				MerchantHomeRegion:     "russia_and_cis",
				PayerRegion:            "europe",
			},
			{
				MinAmount:              5,
				MaxAmount:              999999999.99,
				MethodName:             "MasterCard",
				MethodPercentFee:       1.8,
				MethodFixedFee:         0.2,
				MethodFixedFeeCurrency: "USD",
				PsPercentFee:           3.0,
				PsFixedFee:             0.3,
				PsFixedFeeCurrency:     "USD",
				MerchantHomeRegion:     "russia_and_cis",
				PayerRegion:            "europe",
			},
		},
		reporterpb.RequestParameterAgreementHomeRegion:                         "CIS",
		reporterpb.RequestParameterAgreementMerchantAuthorizedName:             "Test Unit",
		reporterpb.RequestParameterAgreementMerchantAuthorizedPosition:         "Unit test",
		reporterpb.RequestParameterAgreementOperatingCompanyLegalName:          "Unit test",
		reporterpb.RequestParameterAgreementOperatingCompanyAddress:            "Unit test",
		reporterpb.RequestParameterAgreementOperatingCompanyRegistrationNumber: "Unit test",
		reporterpb.RequestParameterAgreementOperatingCompanyAuthorizedName:     "Unit test",
		reporterpb.RequestParameterAgreementOperatingCompanyAuthorizedPosition: "Unit test",
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &reporterpb.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.NoError(suite.T(), err)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_GetParams_Error() {
	handler := &Handler{report: &reporterpb.ReportFile{Params: []byte("\nnot_json_string\n")}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err := builder.Validate()
	assert.Error(suite.T(), err)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_ParamsWithoutRequiredField_Error() {
	params := map[string]interface{}{
		reporterpb.RequestParameterAgreementNumber: "123-TEST",
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &reporterpb.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(errorRequestParameterIsRequired, reporterpb.RequestParameterAgreementLegalName), err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_StringParamIsEmpty_Error() {
	params := map[string]interface{}{
		reporterpb.RequestParameterAgreementNumber: "",
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &reporterpb.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(errorRequestParameterIsEmpty, reporterpb.RequestParameterAgreementNumber), err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_NumericParamIsEmpty_Error() {
	params := map[string]interface{}{
		reporterpb.RequestParameterAgreementNumber:             "123-TEST",
		reporterpb.RequestParameterAgreementLegalName:          "Test",
		reporterpb.RequestParameterAgreementAddress:            "Russia, St.Petersburg, Unit Test 1",
		reporterpb.RequestParameterAgreementRegistrationNumber: "0000000000000000000000001",
		reporterpb.RequestParameterAgreementPayoutCost:         int32(0),
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &reporterpb.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(errorRequestParameterIsEmpty, reporterpb.RequestParameterAgreementPayoutCost), err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Build_Ok() {
	body := []byte(`
		{
		  "number": "0000001",
		  "ps_rate": [
			{
			  "min_amount": 0,
			  "max_amount": 4.99,
			  "method_name": "VISA",
			  "method_percent_fee": 1.8,
			  "method_fixed_fee": 0.2,
			  "method_fixed_fee_currency": "USD",
			  "ps_percent_fee": 3.0,
			  "ps_fixed_fee": 0.3,
			  "ps_fixed_fee_currency": "USD",
			  "merchant_home_region": "europe",
			  "payer_region": "europe"
			},
			{
			  "min_amount": 0,
			  "max_amount": 4.99,
			  "method_name": "MasterCard",
			  "method_percent_fee": 1.8,
			  "method_fixed_fee": 0.2,
			  "method_fixed_fee_currency": "USD",
			  "ps_percent_fee": 3.0,
			  "ps_fixed_fee": 0.3,
			  "ps_fixed_fee_currency": "USD",
			  "merchant_home_region": "europe",
			  "payer_region": "europe"
			},
			{
			  "min_amount": 0,
			  "max_amount": 99999999,
			  "method_name": "Union Pay",
			  "method_percent_fee": 1.8,
			  "method_fixed_fee": 0.2,
			  "method_fixed_fee_currency": "USD",
			  "ps_percent_fee": 3.0,
			  "ps_fixed_fee": 0.3,
			  "ps_fixed_fee_currency": "USD",
			  "merchant_home_region": "europe",
			  "payer_region": "europe"
			}
		  ]
		}`)

	handler := &Handler{report: &reporterpb.ReportFile{Params: body}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	params, err := builder.Build()
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), params, reporterpb.RequestParameterAgreementNumber)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_PostProcess_Ok() {
	bs := &billingMocks.BillingService{}
	bs.On("SetMerchantS3Agreement", mock.Anything, mock.Anything, mock.Anything).
		Return(&billingpb.ChangeMerchantDataResponse{Status: billingpb.ResponseStatusOk}, nil)

	handler := &Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff"},
		billing: bs,
	}
	builder := newAgreementHandler(handler)
	err := builder.PostProcess(context.TODO(), "id", "fileName", 3600, []byte{})
	assert.NoError(suite.T(), err)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_PostProcess_BillingServerSystem_Error() {
	bs := &billingMocks.BillingService{}
	bs.On("SetMerchantS3Agreement", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("some error"))

	handler := &Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff"},
		billing: bs,
	}
	builder := newAgreementHandler(handler)
	err := builder.PostProcess(context.TODO(), "id", "fileName", 3600, []byte{})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "some error", err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_PostProcess_BillingServerReturn_Error() {
	bs := &billingMocks.BillingService{}
	bs.On("SetMerchantS3Agreement", mock.Anything, mock.Anything, mock.Anything).
		Return(
			&billingpb.ChangeMerchantDataResponse{
				Status:  billingpb.ResponseStatusBadData,
				Message: &billingpb.ResponseErrorMessage{Message: "some business logic  error"},
			},
			nil,
		)

	handler := &Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff"},
		billing: bs,
	}
	builder := newAgreementHandler(handler)
	err := builder.PostProcess(context.TODO(), "id", "fileName", 3600, []byte{})
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "some business logic  error", err.Error())
}
