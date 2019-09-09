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

type VatTransactionsBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_VatTransactionsBuilder(t *testing.T) {
	suite.Run(t, new(VatTransactionsBuilderTestSuite))
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Validate_Error_IdNotFound() {
	h := newVatTransactionsHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Validate_Ok() {
	h := newVatTransactionsHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{
			pkg.ParamsFieldId: 1,
		}},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetById() {
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	h := newVatTransactionsHandler(&Handler{
		vatReportRepository: &vatRep,
		report:              &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Error_GetByVat() {
	report := &billingProto.MgoVatReport{Id: bson.NewObjectId()}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByVat", mock2.Anything).Return(nil, errs.New("not found"))

	h := newVatTransactionsHandler(&Handler{
		vatReportRepository:    &vatRep,
		transactionsRepository: &transRep,
		report:                 &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatTransactionsBuilderTestSuite) TestVatTransactionsBuilder_Build_Ok() {
	report := &billingProto.MgoVatReport{Id: bson.NewObjectId()}
	orders := []*billingProto.MgoOrderViewPublic{{Id: bson.NewObjectId()}}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(report, nil)
	transRep := mocks.TransactionsRepositoryInterface{}
	transRep.On("GetByVat", report).Return(orders, nil)

	h := newVatTransactionsHandler(&Handler{
		vatReportRepository:    &vatRep,
		transactionsRepository: &transRep,
		report:                 &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), r, 1)
}
