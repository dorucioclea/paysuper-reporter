package builder

import (
	billingMocks "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type BuilderTestSuite struct {
	suite.Suite
}

func Test_Builder(t *testing.T) {
	suite.Run(t, new(BuilderTestSuite))
}

func (suite *BuilderTestSuite) TestBuilder_NewBuilder_Ok() {
	builder := NewBuilder(
		nil,
		&reporterpb.ReportFile{},
		&billingMocks.BillingService{},
	)

	assert.IsType(suite.T(), &Handler{}, builder)
}

func (suite *BuilderTestSuite) TestBuilder_GetBuilder_Error_NotFound() {
	builder := NewBuilder(
		nil,
		&reporterpb.ReportFile{ReportType: "unknown"},
		&billingMocks.BillingService{},
	)
	_, err := builder.GetBuilder()

	assert.Errorf(suite.T(), err, errors.ErrorHandlerNotFound.Message)
}

func (suite *BuilderTestSuite) TestBuilder_GetBuilder_Ok() {
	builder := NewBuilder(
		nil,
		&reporterpb.ReportFile{ReportType: pkg.ReportTypeVat},
		&billingMocks.BillingService{},
	)
	bldr, err := builder.GetBuilder()

	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &Vat{}, bldr)
}
