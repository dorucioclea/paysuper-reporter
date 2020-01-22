package builder

import (
	"encoding/json"
	errs "errors"
	billPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"testing"
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
		report: &proto.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: "5ced34d689fce60bf4440829",
	})
	h := newVatTransactionsHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetById() {
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		vatRepository: &vatRep,
		report:        &proto.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetByVat() {
	report := &billingProto.MgoVatReport{Id: primitive.NewObjectID()}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByVat", mock2.Anything).Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		vatRepository:          &vatRep,
		transactionsRepository: &transRep,
		report:                 &proto.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Ok() {
	bs := &billMocks.BillingService{}
	response := &grpc.GetOperatingCompanyResponse{
		Status: billPkg.ResponseStatusOk,
		Company: &billingProto.OperatingCompany{
			Name:      "Name",
			Address:   "Address",
			VatNumber: "VatNumber",
		},
	}
	bs.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	report := &billingProto.MgoVatReport{Id: primitive.NewObjectID()}
	orders := []*billingProto.MgoOrderViewPrivate{{
		Id: primitive.NewObjectID(),
		PaymentMethod: &billingProto.MgoOrderPaymentMethod{
			Name: "card",
		},
		TaxFeeTotal: &billingProto.OrderViewMoney{
			Amount: float64(1),
		},
		FeesTotal: &billingProto.OrderViewMoney{
			Amount: float64(1),
		},
		GrossRevenue: &billingProto.OrderViewMoney{
			Amount: float64(1),
		},
	}}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByVat", report).Return(orders, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		vatRepository:          &vatRep,
		transactionsRepository: &transRep,
		report:                 &proto.ReportFile{Params: params},
		billing:                bs,
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetOperatingCompany() {
	bs := &billMocks.BillingService{}
	response := &grpc.GetOperatingCompanyResponse{
		Status:  billPkg.ResponseStatusBadData,
		Message: &grpc.ResponseErrorMessage{Message: "some business logic error"},
	}
	bs.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	report := &billingProto.MgoVatReport{Id: primitive.NewObjectID()}
	orders := []*billingProto.MgoOrderViewPrivate{{
		Id: primitive.NewObjectID(),
		PaymentMethod: &billingProto.MgoOrderPaymentMethod{
			Name: "card",
		},
		TaxFeeTotal: &billingProto.OrderViewMoney{
			Amount: float64(1),
		},
		FeesTotal: &billingProto.OrderViewMoney{
			Amount: float64(1),
		},
		GrossRevenue: &billingProto.OrderViewMoney{
			Amount: float64(1),
		},
	}}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByVat", report).Return(orders, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatTransactionsHandler(&Handler{
		vatRepository:          &vatRep,
		transactionsRepository: &transRep,
		report:                 &proto.ReportFile{Params: params},
		billing:                bs,
	})

	_, err := h.Build()
	assert.Errorf(suite.T(), err, "some business logic error")
}
