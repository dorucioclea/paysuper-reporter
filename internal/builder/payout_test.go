package builder

import (
	"encoding/json"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	billingMocks "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type PayoutBuilderTestSuite struct {
	suite.Suite
}

func Test_PayoutBuilder(t *testing.T) {
	suite.Run(t, new(PayoutBuilderTestSuite))
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Validate_Error_IdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		reporterpb.ParamsFieldId:         "ffffffffffffffffffffffff",
		reporterpb.ParamsFieldMerchantId: "ffffffffffffffffffffffff",
	})
	h := newPayoutHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Ok() {
	billing := &billingMocks.BillingService{}

	payoutResponse := &billingpb.PayoutDocumentResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getPayoutDocumentTemplate(),
	}
	billing.On("GetPayoutDocument", mock2.Anything, mock2.Anything).Return(payoutResponse, nil)

	merchantResponse := &billingpb.GetMerchantResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getMerchantTemplate(),
	}
	billing.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(merchantResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), payoutResponse.Item.Id, r)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Error_GetPayoutDocument() {
	billing := &billingMocks.BillingService{}

	payoutResponse := &billingpb.PayoutDocumentResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Item:    nil,
	}
	billing.On("GetPayoutDocument", mock2.Anything, mock2.Anything).Return(payoutResponse, nil)

	merchantResponse := &billingpb.GetMerchantResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getMerchantTemplate(),
	}
	billing.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(merchantResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Error_GetMerchantBy() {
	billing := &billingMocks.BillingService{}

	payoutResponse := &billingpb.PayoutDocumentResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getPayoutDocumentTemplate(),
	}
	billing.On("GetPayoutDocument", mock2.Anything, mock2.Anything).Return(payoutResponse, nil)

	merchantResponse := &billingpb.GetMerchantResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Item:    nil,
	}
	billing.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(merchantResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Error_GetOperatingCompany() {
	billing := &billingMocks.BillingService{}

	payoutResponse := &billingpb.PayoutDocumentResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getPayoutDocumentTemplate(),
	}
	billing.On("GetPayoutDocument", mock2.Anything, mock2.Anything).Return(payoutResponse, nil)

	merchantResponse := &billingpb.GetMerchantResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getMerchantTemplate(),
	}
	billing.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(merchantResponse, nil)

	response := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusBadData,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Company: nil,
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *PayoutBuilderTestSuite) getPayoutDocumentTemplate() *billingpb.PayoutDocument {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return &billingpb.PayoutDocument{
		Id: "ffffffffffffffffffffffff",
		Destination: &billingpb.MerchantBanking{
			Address: "",
			Details: "",
		},
		Company: &billingpb.MerchantCompanyInfo{
			TaxId: "",
		},
		CreatedAt:  datetime,
		PeriodFrom: datetime,
		PeriodTo:   datetime,
	}
}

func (suite *PayoutBuilderTestSuite) getMerchantTemplate() *billingpb.Merchant {
	return &billingpb.Merchant{
		Id:      "ffffffffffffffffffffffff",
		Company: &billingpb.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
}

func (suite *PayoutBuilderTestSuite) getOperatingCompanyTemplate() *billingpb.OperatingCompany {
	return &billingpb.OperatingCompany{
		Name:      "Name",
		Address:   "Address",
		VatNumber: "VatNumber",
	}
}
