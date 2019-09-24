package builder

import (
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
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
		&proto.ReportFile{},
		&mocks.RoyaltyRepositoryInterface{},
		&mocks.VatRepositoryInterface{},
		&mocks.TransactionsRepositoryInterface{},
		&mocks.PayoutRepositoryInterface{},
	)

	assert.IsType(suite.T(), &Handler{}, builder)
}

func (suite *BuilderTestSuite) TestBuilder_GetBuilder_Error_NotFound() {
	builder := NewBuilder(
		nil,
		&proto.ReportFile{ReportType: "unknown"},
		&mocks.RoyaltyRepositoryInterface{},
		&mocks.VatRepositoryInterface{},
		&mocks.TransactionsRepositoryInterface{},
		&mocks.PayoutRepositoryInterface{},
	)
	_, err := builder.GetBuilder()

	assert.Errorf(suite.T(), err, errors.ErrorHandlerNotFound.Message)
}

func (suite *BuilderTestSuite) TestBuilder_GetBuilder_Ok() {
	builder := NewBuilder(
		nil,
		&proto.ReportFile{ReportType: pkg.ReportTypeVat},
		&mocks.RoyaltyRepositoryInterface{},
		&mocks.VatRepositoryInterface{},
		&mocks.TransactionsRepositoryInterface{},
		&mocks.PayoutRepositoryInterface{},
	)
	bldr, err := builder.GetBuilder()

	assert.NoError(suite.T(), err)
	assert.IsType(suite.T(), &Vat{}, bldr)
}
