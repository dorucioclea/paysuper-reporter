package repository

import (
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
)

type RoyaltyRepositoryTestSuite struct {
	suite.Suite
	db      *mongodb.Source
	service RoyaltyRepositoryInterface
	log     *zap.Logger
}

func Test_RoyaltyRepository(t *testing.T) {
	suite.Run(t, new(RoyaltyRepositoryTestSuite))
}

func (suite *RoyaltyRepositoryTestSuite) SetupTest() {
	var err error

	suite.db, err = mongodb.NewDatabase()

	if err != nil {
		suite.FailNow("Database connection failed", "%v", err)
	}

	suite.log, err = zap.NewProduction()

	if err != nil {
		suite.FailNow("Logger initialization failed", "%v", err)
	}

	suite.service = NewRoyaltyReportRepository(suite.db)
}

func (suite *RoyaltyRepositoryTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	suite.db.Close()
}

func (suite *RoyaltyRepositoryTestSuite) TestRoyaltyRepository_Insert_Error() {
	report := &billingProto.MgoRoyaltyReport{}
	err := suite.service.Insert(report)
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyRepositoryTestSuite) TestRoyaltyRepository_Insert_Ok() {
	report := &billingProto.MgoRoyaltyReport{Id: bson.NewObjectId(), MerchantId: bson.NewObjectId()}
	err := suite.service.Insert(report)
	assert.NoError(suite.T(), err, "unable to insert the royalty report")
}

func (suite *RoyaltyRepositoryTestSuite) TestRoyaltyRepository_GetById_Error() {
	_, err := suite.service.GetById(bson.NewObjectId().Hex())
	assert.Error(suite.T(), err)
}

func (suite *RoyaltyRepositoryTestSuite) TestRoyaltyRepository_GetById_Ok() {
	report := &billingProto.MgoRoyaltyReport{Id: bson.NewObjectId(), MerchantId: bson.NewObjectId()}
	assert.NoError(suite.T(), suite.service.Insert(report))

	rep, err := suite.service.GetById(report.Id.Hex())
	assert.NoError(suite.T(), err, "unable to get the royalty report")
	assert.Equal(suite.T(), report.Id, rep.Id)
	assert.Equal(suite.T(), report.MerchantId, rep.MerchantId)
}
