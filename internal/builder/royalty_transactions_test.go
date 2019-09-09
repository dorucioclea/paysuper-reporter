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

type RoyaltyTransactionsBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_RoyaltyTransactionsBuilder(t *testing.T) {
	suite.Run(t, new(RoyaltyTransactionsBuilderTestSuite))
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Error_IdNotFound() {
	h := newRoyaltyHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Validate_Ok() {
	h := newRoyaltyHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{
			pkg.ParamsFieldId: 1,
		}},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetById() {
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	h := newRoyaltyHandler(&Handler{
		royaltyReportRepository: &royaltyRep,
		report:                  &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestRoyaltyTransactionsBuilder_Build_Error_GetByVat() {
	report := &billingProto.MgoRoyaltyReport{Id: bson.NewObjectId()}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByRoyalty", mock2.Anything).Return(nil, errs.New("not found"))

	h := newRoyaltyTransactionsHandler(&Handler{
		royaltyReportRepository: &royaltyRep,
		transactionsRepository:  &transRep,
		report:                  &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Ok() {
	report := &billingProto.MgoRoyaltyReport{Id: bson.NewObjectId()}
	orders := []*billingProto.MgoOrderViewPublic{{Id: bson.NewObjectId()}}
	royaltyRep := mocks.RoyaltyRepositoryInterface{}
	royaltyRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByRoyalty", report).Return(orders, nil)

	h := newRoyaltyTransactionsHandler(&Handler{
		royaltyReportRepository: &royaltyRep,
		transactionsRepository:  &transRep,
		report:                  &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), r, 1)
}
