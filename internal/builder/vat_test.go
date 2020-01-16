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
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Error_CountryInvalid() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "ABC",
	})
	h := newVatHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Ok() {
	billing := &billingMocks.BillingService{}

	reportsResponse := &billingpb.VatReportsResponse{
		Status: billingpb.ResponseStatusOk,
		Data: &billingpb.VatReportsPaginate{
			Items: suite.getReportsTemplate(),
		},
	}
	billing.On("GetVatReportsForCountry", mock2.Anything, mock2.Anything).Return(reportsResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	response := &billingpb.GetOperatingCompanyResponse{
		Status: billingpb.ResponseStatusOk,
		Company: &billingpb.OperatingCompany{
			Name:      "Name",
			Address:   "Address",
			VatNumber: "VatNumber",
		},
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), reportsResponse.Data.Items, 1)
	assert.NotEmpty(suite.T(), reportsResponse.Data.Items[0].Id, r)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Error_GetVatReportsForCountry() {
	billing := &billingMocks.BillingService{}

	reportsResponse := &billingpb.VatReportsResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Data: &billingpb.VatReportsPaginate{
			Items: nil,
		},
	}
	billing.On("GetVatReportsForCountry", mock2.Anything, mock2.Anything).Return(reportsResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusOk,
		Company: suite.getOperatingCompanyTemplate(),
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	response := &billingpb.GetOperatingCompanyResponse{
		Status: billingpb.ResponseStatusOk,
		Company: &billingpb.OperatingCompany{
			Name:      "Name",
			Address:   "Address",
			VatNumber: "VatNumber",
		},
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Error_GetOperatingCompany() {
	billing := &billingMocks.BillingService{}

	reportsResponse := &billingpb.VatReportsResponse{
		Status: billingpb.ResponseStatusOk,
		Data: &billingpb.VatReportsPaginate{
			Items: suite.getReportsTemplate(),
		},
	}
	billing.On("GetVatReportsForCountry", mock2.Anything, mock2.Anything).Return(reportsResponse, nil)

	ocResponse := &billingpb.GetOperatingCompanyResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Company: nil,
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(ocResponse, nil)

	response := &billingpb.GetOperatingCompanyResponse{
		Status: billingpb.ResponseStatusOk,
		Company: &billingpb.OperatingCompany{
			Name:      "Name",
			Address:   "Address",
			VatNumber: "VatNumber",
		},
	}
	billing.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldCountry: "RU",
	})
	h := newVatHandler(&Handler{
		report:  &reporterpb.ReportFile{Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatBuilderTestSuite) getReportsTemplate() []*billingpb.VatReport {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return []*billingpb.VatReport{
		{
			Id:           bson.NewObjectId().Hex(),
			DateFrom:     datetime,
			DateTo:       datetime,
			PayUntilDate: datetime,
		},
	}
}

func (suite *VatBuilderTestSuite) getOperatingCompanyTemplate() *billingpb.OperatingCompany {
	return &billingpb.OperatingCompany{
		Name:      "Name",
		Address:   "Address",
		VatNumber: "VatNumber",
	}
}
