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

type VatTransactionsBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_VatTransactionsBuilder(t *testing.T) {
	suite.Run(t, new(VatTransactionsBuilderTestSuite))
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Validate_Error_IdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: bson.NewObjectId().Hex(),
	})
	h := newVatTransactionsHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Ok() {
	billing := &billingMocks.BillingService{}

	vatResponse := &billingpb.VatReportResponse{
		Status: billingpb.ResponseStatusOk,
		Vat:    suite.getVatTemplate(),
	}
	billing.On("GetVatReport", mock2.Anything, mock2.Anything).Return(vatResponse, nil)

	ordersResponse := &billingpb.TransactionsResponse{
		Status: billingpb.ResponseStatusOk,
		Data: &billingpb.TransactionsPaginate{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("GetVatReportTransactions", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetVatReport() {
	billing := &billingMocks.BillingService{}

	vatResponse := &billingpb.VatReportResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Vat:     nil,
	}
	billing.On("GetVatReport", mock2.Anything, mock2.Anything).Return(vatResponse, nil)

	ordersResponse := &billingpb.TransactionsResponse{
		Status: billingpb.ResponseStatusOk,
		Data: &billingpb.TransactionsPaginate{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("GetVatReportTransactions", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetVatReportTransactions() {
	billing := &billingMocks.BillingService{}

	vatResponse := &billingpb.VatReportResponse{
		Status: billingpb.ResponseStatusOk,
		Vat:    suite.getVatTemplate(),
	}
	billing.On("GetVatReport", mock2.Anything, mock2.Anything).Return(vatResponse, nil)

	ordersResponse := &billingpb.TransactionsResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Data: &billingpb.TransactionsPaginate{
			Items: nil,
		},
	}
	billing.On("GetVatReportTransactions", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetOperatingCompany() {
	billing := &billingMocks.BillingService{}

	vatResponse := &billingpb.VatReportResponse{
		Status: billingpb.ResponseStatusOk,
		Vat:    suite.getVatTemplate(),
	}
	billing.On("GetVatReport", mock2.Anything, mock2.Anything).Return(vatResponse, nil)

	ordersResponse := &billingpb.TransactionsResponse{
		Status: billingpb.ResponseStatusOk,
		Data: &billingpb.TransactionsPaginate{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("GetVatReportTransactions", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Company: nil,
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) getVatTemplate() *billingpb.VatReport {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return &billingpb.VatReport{
		Id:        bson.NewObjectId().Hex(),
		CreatedAt: datetime,
		DateFrom:  datetime,
		DateTo:    datetime,
	}
}

func (suite *VatTransactionsBuilderTestSuite) getOrdersTemplate() []*billingpb.OrderViewPublic {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return []*billingpb.OrderViewPublic{{
		Id: bson.NewObjectId().Hex(),
		PaymentMethod: &billingpb.PaymentMethodOrder{
			Name: "card",
		},
		TaxFeeTotal: &billingpb.OrderViewMoney{
			Amount: float64(1),
		},
		FeesTotal: &billingpb.OrderViewMoney{
			Amount: float64(1),
		},
		GrossRevenue: &billingpb.OrderViewMoney{
			Amount: float64(1),
		},
		TransactionDate: datetime,
	}}
}

func (suite *VatTransactionsBuilderTestSuite) getOperatingCompanyTemplate() *billingpb.OperatingCompany {
	return &billingpb.OperatingCompany{
		Name:      "Name",
		Address:   "Address",
		VatNumber: "VatNumber",
	}
}
