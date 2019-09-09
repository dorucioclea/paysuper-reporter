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

type VatBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_VatBuilder(t *testing.T) {
	suite.Run(t, new(VatBuilderTestSuite))
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Error_IdNotFound() {
	h := newVatHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamIdNotFound.Message)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Validate_Ok() {
	h := newVatHandler(&Handler{
		report: &proto.MgoReportFile{Params: map[string]interface{}{
			pkg.ParamsFieldId: 1,
		}},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Error_GetById() {
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(nil, errs.New("not found"))

	h := newVatHandler(&Handler{
		vatReportRepository: &vatRep,
		report:              &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *VatBuilderTestSuite) TestVatBuilder_Build_Ok() {
	report := &billingProto.MgoVatReport{Id: bson.NewObjectId()}
	vatRep := mocks.VatRepositoryInterface{}
	vatRep.On("GetById", mock2.Anything).Return(report, nil)

	h := newVatHandler(&Handler{
		vatReportRepository: &vatRep,
		report:              &proto.MgoReportFile{Params: map[string]interface{}{}},
	})

	r, err := h.Build()
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), report.Id, r)
}
