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
)

type PayoutBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
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
	report := &billingProto.MgoPayoutDocument{
		Id: bson.NewObjectId(),
		Destination: &billingProto.MerchantBanking{
			Address: "",
			Details: "",
		},
		Summary: &billingProto.PayoutDocumentSummary{
			Orders: &billingProto.PayoutDocumentOrders{
				Count: 0,
			},
		},
	}
	payoutRep := mocks.PayoutRepositoryInterface{}
	payoutRep.On("GetById", mock2.Anything).Return(report, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newPayoutHandler(&Handler{
		payoutRepository: &payoutRep,
		report:           &proto.ReportFile{Params: params},
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), report.Id, r)
}
