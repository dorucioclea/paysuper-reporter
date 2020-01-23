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

type RoyaltyTransactionsBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_RoyaltyTransactionsBuilder(t *testing.T) {
	suite.Run(t, new(RoyaltyTransactionsBuilderTestSuite))
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Error_MerchantIdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{
		reporterpb.ParamsFieldId: "ffffffffffffffffffffffff",
	})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Error_IdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		reporterpb.ParamsFieldId: "ffffffffffffffffffffffff",
	})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Ok() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status: billingpb.ResponseStatusOk,
		Item: &billingpb.ListOrdersPublicResponseItem{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

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
	h := newRoyaltyTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetRoyaltyReport() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Item:    nil,
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status: billingpb.ResponseStatusOk,
		Item: &billingpb.ListOrdersPublicResponseItem{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

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
	h := newRoyaltyTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_FindAllOrdersPublic() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Item: &billingpb.ListOrdersPublicResponseItem{
			Items: nil,
		},
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

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
	h := newRoyaltyTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetMerchantBy() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status: billingpb.ResponseStatusOk,
		Item: &billingpb.ListOrdersPublicResponseItem{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

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
	h := newRoyaltyTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetOperatingCompany() {
	billing := &billingMocks.BillingService{}

	royaltyResponse := &billingpb.GetRoyaltyReportResponse{
		Status: billingpb.ResponseStatusOk,
		Item:   suite.getRoyaltyReportTemplate(),
	}
	billing.On("GetRoyaltyReport", mock2.Anything, mock2.Anything).Return(royaltyResponse, nil)

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status: billingpb.ResponseStatusOk,
		Item: &billingpb.ListOrdersPublicResponseItem{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

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
	h := newRoyaltyTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: "ffffffffffffffffffffffff", Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) getRoyaltyReportTemplate() *billingpb.RoyaltyReport {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return &billingpb.RoyaltyReport{
		Id:         "ffffffffffffffffffffffff",
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

func (suite *RoyaltyTransactionsBuilderTestSuite) getOrdersTemplate() []*billingpb.OrderViewPublic {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return []*billingpb.OrderViewPublic{{
		Id:          "ffffffffffffffffffffffff",
		Transaction: "1",
		CountryCode: "RU",
		Currency:    "RUB",
		Project:     &billingpb.ProjectOrder{Name: map[string]string{"en": "name"}},
		PaymentMethod: &billingpb.PaymentMethodOrder{
			Name: "card",
		},
		CreatedAt:       datetime,
		TransactionDate: datetime,
		NetRevenue: &billingpb.OrderViewMoney{
			Amount: 1,
		},
	}}
}

func (suite *RoyaltyTransactionsBuilderTestSuite) getMerchantTemplate() *billingpb.Merchant {
	return &billingpb.Merchant{
		Id:      "ffffffffffffffffffffffff",
		Company: &billingpb.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
}

func (suite *RoyaltyTransactionsBuilderTestSuite) getOperatingCompanyTemplate() *billingpb.OperatingCompany {
	return &billingpb.OperatingCompany{
		Name:      "Name",
		Address:   "Address",
		VatNumber: "VatNumber",
	}
}
