package repository

import (
	"fmt"
	"github.com/globalsign/mgo/bson"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
)

type VatRepositoryTestSuite struct {
	suite.Suite
	db      *mongodb.Source
	service VatRepositoryInterface
	log     *zap.Logger
}

func Test_VatRepository(t *testing.T) {
	suite.Run(t, new(VatRepositoryTestSuite))
}

func (suite *VatRepositoryTestSuite) SetupTest() {
	var err error

	suite.db, err = mongodb.NewDatabase()

	if err != nil {
		suite.FailNow("Database connection failed", "%v", err)
	}

	suite.log, err = zap.NewProduction()

	if err != nil {
		suite.FailNow("Logger initialization failed", "%v", err)
	}

	suite.service = NewVatRepository(suite.db)
}

func (suite *VatRepositoryTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	suite.db.Close()
}

func (suite *VatRepositoryTestSuite) TestVatRepository_Insert_Error() {
	report := &billingProto.MgoVatReport{}
	err := suite.service.Insert(report)
	assert.Error(suite.T(), err)
}

func (suite *VatRepositoryTestSuite) TestVatRepository_Insert_Ok() {
	report := &billingProto.MgoVatReport{Id: bson.NewObjectId()}
	err := suite.service.Insert(report)
	assert.NoError(suite.T(), err, "unable to insert the vat report")
}

func (suite *VatRepositoryTestSuite) TestVatRepository_GetById_Error() {
	_, err := suite.service.GetById(bson.NewObjectId().Hex())
	assert.Error(suite.T(), err)
}

func (suite *VatRepositoryTestSuite) TestVatRepository_GetById_Error_ByStatus() {
	report := &billingProto.MgoVatReport{Id: bson.NewObjectId()}
	assert.NoError(suite.T(), suite.service.Insert(report))

	_, err := suite.service.GetById(report.Id.Hex())
	assert.Error(suite.T(), err)
}

func (suite *VatRepositoryTestSuite) TestVatRepository_GetById_Ok() {
	report := &billingProto.MgoVatReport{Id: bson.NewObjectId(), Status: billingPkg.VatReportStatusNeedToPay}
	assert.NoError(suite.T(), suite.service.Insert(report))

	rep, err := suite.service.GetById(report.Id.Hex())
	fmt.Println(rep)
	assert.NoError(suite.T(), err, "unable to get the vat report")
	assert.Equal(suite.T(), report.Id, rep.Id)
}
