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
	"time"
)

type RoyaltyTransactionsBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_RoyaltyTransactionsBuilder(t *testing.T) {
	suite.Run(t, new(RoyaltyTransactionsBuilderTestSuite))
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Error_IdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: "5ced34d689fce60bf4440829",
	})
	h := newRoyaltyHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetById() {
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		royaltyRepository: &royaltyRep,
		report:            &proto.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetByVat() {
	report := &billingProto.MgoRoyaltyReport{Id: primitive.NewObjectID()}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByRoyalty", mock2.Anything).Return(nil, errs.New("not found"))

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchantRep.
		On("GetById", mock2.Anything).
		Return(
			&billingProto.MgoMerchant{
				Id:      primitive.NewObjectID(),
				Company: &billingProto.MerchantCompanyInfo{Name: "", Address: ""},
			},
			nil,
		)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyTransactionsHandler(&Handler{
		royaltyRepository:      &royaltyRep,
		transactionsRepository: &transRep,
		merchantRepository:     &merchantRep,
		report:                 &proto.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Ok() {
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

	report := &billingProto.MgoRoyaltyReport{Id: primitive.NewObjectID()}
	orders := []*billingProto.MgoOrderViewPublic{{
		Id:          primitive.NewObjectID(),
		Transaction: "1",
		CountryCode: "RU",
		Currency:    "RUB",
		Project:     &billingProto.MgoOrderProject{Name: []*billingProto.MgoMultiLang{{Value: ""}}},
		PaymentMethod: &billingProto.MgoOrderPaymentMethod{
			Name: "card",
		},
		CreatedAt: time.Now(),
		NetRevenue: &billingProto.OrderViewMoney{
			Amount: 1,
		},
	}}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByRoyalty", report).Return(orders, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchantRep.
		On("GetById", mock2.Anything).
		Return(
			&billingProto.MgoMerchant{
				Id:      primitive.NewObjectID(),
				Company: &billingProto.MerchantCompanyInfo{Name: "", Address: ""},
			},
			nil,
		)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyTransactionsHandler(&Handler{
		royaltyRepository:      &royaltyRep,
		transactionsRepository: &transRep,
		merchantRepository:     &merchantRep,
		report:                 &proto.ReportFile{Params: params},
		billing:                bs,
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetOperatingCompany() {
	bs := &billMocks.BillingService{}
	response := &grpc.GetOperatingCompanyResponse{
		Status:  billPkg.ResponseStatusBadData,
		Message: &grpc.ResponseErrorMessage{Message: "some business logic error"},
	}
	bs.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	report := &billingProto.MgoRoyaltyReport{Id: primitive.NewObjectID()}
	orders := []*billingProto.MgoOrderViewPublic{{
		Id:          primitive.NewObjectID(),
		Transaction: "1",
		CountryCode: "RU",
		Currency:    "RUB",
		Project:     &billingProto.MgoOrderProject{Name: []*billingProto.MgoMultiLang{{Value: ""}}},
		PaymentMethod: &billingProto.MgoOrderPaymentMethod{
			Name: "card",
		},
		CreatedAt: time.Now(),
		NetRevenue: &billingProto.OrderViewMoney{
			Amount: 1,
		},
	}}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByRoyalty", report).Return(orders, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchantRep.
		On("GetById", mock2.Anything).
		Return(
			&billingProto.MgoMerchant{
				Id:      primitive.NewObjectID(),
				Company: &billingProto.MerchantCompanyInfo{Name: "", Address: ""},
			},
			nil,
		)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyTransactionsHandler(&Handler{
		royaltyRepository:      &royaltyRep,
		transactionsRepository: &transRep,
		merchantRepository:     &merchantRep,
		report:                 &proto.ReportFile{Params: params},
		billing:                bs,
	})

	_, err := h.Build()
	assert.Errorf(suite.T(), err, "some business logic error")
}
