package builder

import (
	"encoding/json"
	errs "errors"
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
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
		report: &proto.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Ok() {
	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldId: "5ced34d689fce60bf4440829",
	})
	h := newRoyaltyHandler(&Handler{
		report: &proto.ReportFile{Params: params},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetById() {
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

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Ok() {
	datetime := time.Now()
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		PeriodFrom: datetime,
		PeriodTo:   datetime,
		PayoutDate: datetime,
		CreatedAt:  datetime,
		AcceptedAt: datetime,
		Totals: &billingProto.RoyaltyReportTotals{
			VatAmount:            1,
			TransactionsCount:    1,
			RollingReserveAmount: 1,
			FeeAmount:            1,
			CorrectionAmount:     1,
			PayoutAmount:         1,
		},
	}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newRoyaltyHandler(&Handler{
		royaltyRepository: &royaltyRep,
		report:            &proto.ReportFile{Params: params},
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), report.Id, r)
}
