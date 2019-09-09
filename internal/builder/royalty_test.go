package builder

import (
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

type RoyaltyBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_RoyaltyBuilder(t *testing.T) {
	suite.Run(t, new(RoyaltyBuilderTestSuite))
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Error_IdNotFound() {
	h := newRoyaltyHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Validate_Ok() {
	h := newRoyaltyHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{
			pkg.ParamsFieldId: 1,
		}},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Error_GetById() {
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	h := newRoyaltyHandler(&Handler{
		royaltyReportRepository: &royaltyRep,
		report:                  &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyBuilderTestSuite) TestRoyaltyBuilder_Build_Ok() {
	report := &billingProto.MgoRoyaltyReport{Id: bson.NewObjectId()}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)

	h := newRoyaltyHandler(&Handler{
		royaltyReportRepository: &royaltyRep,
		report:                  &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), report.Id, r)
}
