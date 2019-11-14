package builder

import (
	"encoding/json"
	errs "errors"
	"github.com/globalsign/mgo/bson"
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
	"testing"
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
		report: &proto.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: "5ced34d689fce60bf4440829",
	})
	h := newPayoutHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Error_GetById() {
	payoutRep := mocks.PayoutRepositoryInterface{}
	payoutRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		payoutRepository: &payoutRep,
		report:           &proto.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Ok() {
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

	report := &billingProto.MgoPayoutDocument{
		Id: bson.NewObjectId(),
		Destination: &billingProto.MerchantBanking{
			Address: "",
			Details: "",
		},
		Company: &billingProto.MerchantCompanyInfo{
			TaxId: "",
		},
	}
	payoutRep := mocks.PayoutRepositoryInterface{}
	payoutRep.On("GetById", mock2.Anything).Return(report, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchant := &billingProto.MgoMerchant{
		Id:      bson.NewObjectId(),
		Company: &billingProto.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
	merchantRep.On("GetById", mock2.Anything).Return(merchant, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		payoutRepository:   &payoutRep,
		merchantRepository: &merchantRep,
		report:             &proto.ReportFile{Params: params},
		billing:            bs,
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), report.Id, r)
}

func (suite *PayoutBuilderTestSuite) TestPayoutBuilder_Build_Error_GetOperatingCompany() {
	bs := &billMocks.BillingService{}
	response := &grpc.GetOperatingCompanyResponse{
		Status:  billPkg.ResponseStatusBadData,
		Message: &grpc.ResponseErrorMessage{Message: "some business logic error"},
	}
	bs.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	report := &billingProto.MgoPayoutDocument{
		Id: bson.NewObjectId(),
		Destination: &billingProto.MerchantBanking{
			Address: "",
			Details: "",
		},
		Company: &billingProto.MerchantCompanyInfo{
			TaxId: "",
		},
	}
	payoutRep := mocks.PayoutRepositoryInterface{}
	payoutRep.On("GetById", mock2.Anything).Return(report, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchant := &billingProto.MgoMerchant{
		Id:      bson.NewObjectId(),
		Company: &billingProto.MerchantCompanyInfo{Name: "", Address: "", TaxId: ""},
	}
	merchantRep.On("GetById", mock2.Anything).Return(merchant, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		payoutRepository:   &payoutRep,
		merchantRepository: &merchantRep,
		report:             &proto.ReportFile{Params: params},
		billing:            bs,
	})

	_, err := h.Build()
	assert.Errorf(suite.T(), err, "some business logic error")
}
