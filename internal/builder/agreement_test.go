package builder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-micro"
	billPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	billProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AgreementBuilderTestSuite struct {
	suite.Suite
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
		pkg.RequestParameterAgreementNumber:             "123-TEST",
		pkg.RequestParameterAgreementLegalName:          "Test",
		pkg.RequestParameterAgreementAddress:            "Russia, St.Petersburg, Unit Test 1",
		pkg.RequestParameterAgreementRegistrationNumber: "0000000000000000000000001",
		pkg.RequestParameterAgreementPayoutCost:         10,
		pkg.RequestParameterAgreementMinimalPayoutLimit: 10000,
		pkg.RequestParameterAgreementPayoutCurrency:     "USD",
		pkg.RequestParameterAgreementPSRate: []*billProto.MerchantTariffRatesPayment{
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
		pkg.RequestParameterAgreementHomeRegion:                 "CIS",
		pkg.RequestParameterAgreementMerchantAuthorizedName:     "Test Unit",
		pkg.RequestParameterAgreementMerchantAuthorizedPosition: "Unit test",
		pkg.RequestParameterAgreementProjectsLink:               "http://localhost",
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &proto.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.NoError(suite.T(), err)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_GetParams_Error() {
	handler := &Handler{report: &proto.ReportFile{Params: []byte("\nnot_json_string\n")}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err := builder.Validate()
	assert.Error(suite.T(), err)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_ParamsWithoutRequiredField_Error() {
	params := map[string]interface{}{
		pkg.RequestParameterAgreementNumber: "123-TEST",
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &proto.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(errorRequestParameterIsRequired, pkg.RequestParameterAgreementLegalName), err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_StringParamIsEmpty_Error() {
	params := map[string]interface{}{
		pkg.RequestParameterAgreementNumber: "",
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &proto.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(errorRequestParameterIsEmpty, pkg.RequestParameterAgreementNumber), err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_Validate_NumericParamIsEmpty_Error() {
	params := map[string]interface{}{
		pkg.RequestParameterAgreementNumber:             "123-TEST",
		pkg.RequestParameterAgreementLegalName:          "Test",
		pkg.RequestParameterAgreementAddress:            "Russia, St.Petersburg, Unit Test 1",
		pkg.RequestParameterAgreementRegistrationNumber: "0000000000000000000000001",
		pkg.RequestParameterAgreementPayoutCost:         int32(0),
	}
	b, err := json.Marshal(params)
	assert.NoError(suite.T(), err)
	handler := &Handler{report: &proto.ReportFile{Params: b}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	err = builder.Validate()
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), fmt.Sprintf(errorRequestParameterIsEmpty, pkg.RequestParameterAgreementPayoutCost), err.Error())
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

	handler := &Handler{report: &proto.ReportFile{Params: body}, service: micro.NewService()}
	builder := newAgreementHandler(handler)
	params, err := builder.Build()
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), params, pkg.RequestParameterAgreementNumber)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_PostProcess_Ok() {
	bs := &billMocks.BillingService{}
	bs.On("SetMerchantS3Agreement", mock.Anything, mock.Anything, mock.Anything).
		Return(&grpc.ChangeMerchantDataResponse{Status: billPkg.ResponseStatusOk}, nil)

	builder := Agreement{
		Handler: &Handler{
			report:  &proto.ReportFile{MerchantId: bson.NewObjectId().Hex()},
			service: micro.NewService(),
		},
		billingService: bs,
	}
	err := builder.PostProcess(context.TODO(), "id", "fileName", 3600)
	assert.NoError(suite.T(), err)
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_PostProcess_BillingServerSystem_Error() {
	bs := &billMocks.BillingService{}
	bs.On("SetMerchantS3Agreement", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("some error"))

	builder := Agreement{
		Handler: &Handler{
			report:  &proto.ReportFile{MerchantId: bson.NewObjectId().Hex()},
			service: micro.NewService(),
		},
		billingService: bs,
	}
	err := builder.PostProcess(context.TODO(), "id", "fileName", 3600)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "some error", err.Error())
}

func (suite *AgreementBuilderTestSuite) TestAgreementBuilder_PostProcess_BillingServerReturn_Error() {
	bs := &billMocks.BillingService{}
	bs.On("SetMerchantS3Agreement", mock.Anything, mock.Anything, mock.Anything).
		Return(
			&grpc.ChangeMerchantDataResponse{
				Status:  billPkg.ResponseStatusBadData,
				Message: &grpc.ResponseErrorMessage{Message: "some business logic  error"},
			},
			nil,
		)

	builder := Agreement{
		Handler: &Handler{
			report:  &proto.ReportFile{MerchantId: bson.NewObjectId().Hex()},
			service: micro.NewService(),
		},
		billingService: bs,
	}
	err := builder.PostProcess(context.TODO(), "id", "fileName", 3600)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "some business logic  error", err.Error())
}
