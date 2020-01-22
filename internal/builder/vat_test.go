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

type VatBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_VatBuilder(t *testing.T) {
	suite.Run(t, new(VatBuilderTestSuite))
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Error_CountryEmpty() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newVatHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Error_CountryInvalid() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "ABC",
	})
	h := newVatHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Error_GetById() {
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetByCountry", mock2.Anything).Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		vatRepository: &vatRep,
		report:        &proto.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Ok() {
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

	report := []*billingProto.MgoVatReport{{Id: primitive.NewObjectID()}}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetByCountry", mock2.Anything).Return(report, nil)

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		vatRepository: &vatRep,
		report:        &proto.ReportFile{Params: params},
		billing:       bs,
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), report, 1)
	assert.NotEmpty(suite.T(), report[0].Id, r)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Error_GetOperatingCompany() {
	bs := &billMocks.BillingService{}
	response := &grpc.GetOperatingCompanyResponse{
		Status:  billPkg.ResponseStatusBadData,
		Message: &grpc.ResponseErrorMessage{Message: "some business logic error"},
	}
	bs.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	report := []*billingProto.MgoVatReport{{Id: primitive.NewObjectID()}}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetByCountry", mock2.Anything).Return(report, nil)

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		vatRepository: &vatRep,
		report:        &proto.ReportFile{Params: params},
		billing:       bs,
	})

	_, err := h.Build()
	assert.Errorf(suite.T(), err, "some business logic error")
}
