package builder

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	billingMocks "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type RoyaltyBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_RoyaltyBuilder(t *testing.T) {
	suite.Run(t, new(RoyaltyBuilderTestSuite))
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Error_MerchantIdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: bson.NewObjectId().Hex(),
	})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Error_IdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: bson.NewObjectId().Hex(),
	})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Ok() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

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
	h := newRoyaltyHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), royaltyResponse.Item.Id, r)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetRoyaltyReport() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Item:    nil,
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

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
	h := newRoyaltyHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetMerchantBy() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

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
	h := newRoyaltyHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetOperatingCompany() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

	merchantResponse := &billingpb.GetMerchantResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getMerchantTemplate(),
	}
	billing.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(merchantResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Company: nil,
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyBuilderTestSuite) getRoyaltyReportTemplate() *billingpb.RoyaltyReport {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return &billingpb.RoyaltyReport{
		Id:         bson.NewObjectId().Hex(),
		PeriodFrom: datetime,
		PeriodTo:   datetime,
		PayoutDate: datetime,
		CreatedAt:  datetime,
		AcceptedAt: datetime,
		Totals: &billingpb.RoyaltyReportTotals{
			RollingReserveAmount: 1,
			CorrectionAmount:     1,
		},
		Summary: &billingpb.RoyaltyReportSummary{
			ProductsItems: []*billingpb.RoyaltyReportProductSummaryItem{{
				Product:            "",
				Region:             "",
				TotalTransactions:  1,
				ReturnsCount:       1,
				SalesCount:         1,
				GrossSalesAmount:   1,
				GrossReturnsAmount: 1,
				GrossTotalAmount:   1,
				TotalVat:           1,
				TotalFees:          1,
				PayoutAmount:       1,
			}},
			Corrections: nil,
		},
	}
}

func (suite *RoyaltyBuilderTestSuite) getMerchantTemplate() *billingpb.Merchant {
	return &billingpb.Merchant{
		Id:      bson.NewObjectId().String(),
		Company: &billingpb.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
}

func (suite *RoyaltyBuilderTestSuite) getOperatingCompanyTemplate() *billingpb.OperatingCompany {
	return &billingpb.OperatingCompany{
		Name:      "Name",
		Address:   "Address",
		VatNumber: "VatNumber",
	}
}
