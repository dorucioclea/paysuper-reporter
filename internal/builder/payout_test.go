package builder

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	billingMocks "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	errs "errors"
	billPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		reporterpb.ParamsFieldId: bson.NewObjectId().Hex(),
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
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
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
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
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
	report := &billingProto.MgoPayoutDocument{
		Id: primitive.NewObjectID(),
		Destination: &billingProto.MerchantBanking{
			Address: "",
			Details: "",
		},
		Company: &billingProto.MerchantCompanyInfo{
			TaxId: "",
		},
	}
	billing.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(merchantResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	merchantRep := mocks.MerchantRepositoryInterface{}
	merchant := &billingProto.MgoMerchant{
		Id:      primitive.NewObjectID(),
		Company: &billingProto.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
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
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *PayoutBuilderTestSuite) getPayoutDocumentTemplate() *billingpb.PayoutDocument {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return &billingpb.PayoutDocument{
		Id: bson.NewObjectId().Hex(),
		Destination: &billingpb.MerchantBanking{
	report := &billingProto.MgoPayoutDocument{
		Id: primitive.NewObjectID(),
		Destination: &billingProto.MerchantBanking{
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
	payoutRep := mocks.PayoutRepositoryInterface{}
	payoutRep.On("GetById", mock2.Anything).Return(report, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchant := &billingProto.MgoMerchant{
		Id:      primitive.NewObjectID(),
		Company: &billingProto.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
	merchantRep.On("GetById", mock2.Anything).Return(merchant, nil)
}

func (suite *PayoutBuilderTestSuite) getMerchantTemplate() *billingpb.Merchant {
	return &billingpb.Merchant{
		Id:      bson.NewObjectId().String(),
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
