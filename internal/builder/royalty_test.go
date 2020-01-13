package builder

import (
	"encoding/json"
	errs "errors"
	"github.com/globalsign/mgo/bson"
	billPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
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

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Error_IdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: "5ced34d689fce60bf4440829",
	})
	h := newRoyaltyHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetById() {
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		royaltyRepository: &royaltyRep,
		report:            &reporterpb.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Ok() {
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

	datetime := time.Now()
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		PeriodFrom: datetime,
		PeriodTo:   datetime,
		PayoutDate: datetime,
		CreatedAt:  datetime,
		AcceptedAt: datetime,
		Totals: &billingProto.RoyaltyReportTotals{
			RollingReserveAmount: 1,
			CorrectionAmount:     1,
		},
		Summary: &billingProto.RoyaltyReportSummary{
			ProductsItems: []*billingProto.RoyaltyReportProductSummaryItem{{
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
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchantRep.
		On("GetById", mock2.Anything).
		Return(
			&billingProto.MgoMerchant{
				Id:      bson.NewObjectId(),
				Company: &billingProto.MerchantCompanyInfo{Name: "", Address: ""},
			},
			nil,
		)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		royaltyRepository:  &royaltyRep,
		merchantRepository: &merchantRep,
		report:             &reporterpb.ReportFile{Params: params},
		billing:            bs,
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), report.Id, r)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetOperatingCompany() {
	bs := &billMocks.BillingService{}
	response := &grpc.GetOperatingCompanyResponse{
		Status:  billPkg.ResponseStatusBadData,
		Message: &grpc.ResponseErrorMessage{Message: "some business logic error"},
	}
	bs.On("GetOperatingCompany", mock2.Anything, mock2.Anything).Return(response, nil)

	datetime := time.Now()
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		PeriodFrom: datetime,
		PeriodTo:   datetime,
		PayoutDate: datetime,
		CreatedAt:  datetime,
		AcceptedAt: datetime,
		Totals: &billingProto.RoyaltyReportTotals{
			RollingReserveAmount: 1,
			CorrectionAmount:     1,
		},
		Summary: &billingProto.RoyaltyReportSummary{
			ProductsItems: []*billingProto.RoyaltyReportProductSummaryItem{{
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
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	merchantRep := mocks.MerchantRepositoryInterface{}
	merchantRep.
		On("GetById", mock2.Anything).
		Return(
			&billingProto.MgoMerchant{
				Id:      bson.NewObjectId(),
				Company: &billingProto.MerchantCompanyInfo{Name: "", Address: ""},
			},
			nil,
		)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		royaltyRepository:  &royaltyRep,
		merchantRepository: &merchantRep,
		report:             &reporterpb.ReportFile{Params: params},
		billing:            bs,
	})

	_, err := h.Build()
	assert.Errorf(suite.T(), err, "some business logic error")
}
