package repository

import (
	"github.com/globalsign/mgo/bson"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
)

type ReportFileRepositoryTestSuite struct {
	suite.Suite
	db      *mongodb.Source
	service ReportFileRepositoryInterface
	log     *zap.Logger
}

func Test_ReportFileRepository(t *testing.T) {
	suite.Run(t, new(ReportFileRepositoryTestSuite))
}

func (suite *ReportFileRepositoryTestSuite) SetupTest() {
	var err error

	suite.db, err = mongodb.NewDatabase()

	if err != nil {
		suite.FailNow("Database connection failed", "%v", err)
	}

	suite.log, err = zap.NewProduction()

	if err != nil {
		suite.FailNow("Logger initialization failed", "%v", err)
	}

	suite.service = NewReportFileRepository(suite.db)
}

func (suite *ReportFileRepositoryTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	suite.db.Close()
}

func (suite *ReportFileRepositoryTestSuite) TestReportFile_Insert_Error() {
	report := &proto.MgoReportFile{}
	err := suite.service.Insert(report)
	assert.Error(suite.T(), err)
}

func (suite *ReportFileRepositoryTestSuite) TestReportFile_Insert_Ok() {
	report := &proto.MgoReportFile{Id: bson.NewObjectId(), MerchantId: bson.NewObjectId()}
	err := suite.service.Insert(report)
	assert.NoError(suite.T(), err, "unable to insert the report file")
}

func (suite *ReportFileRepositoryTestSuite) TestReportFile_Update_Error() {
	report := &proto.MgoReportFile{}
	err := suite.service.Update(report)
	assert.Error(suite.T(), err)
}

func (suite *ReportFileRepositoryTestSuite) TestReportFile_Update_Ok() {
	report := &proto.MgoReportFile{Id: bson.NewObjectId(), MerchantId: bson.NewObjectId()}
	assert.NoError(suite.T(), suite.service.Insert(report))

	err := suite.service.Update(report)
	assert.NoError(suite.T(), err, "unable to update the report file")
}

func (suite *ReportFileRepositoryTestSuite) TestReportFile_GetById_Error() {
	_, err := suite.service.GetById(bson.NewObjectId().Hex())
	assert.Error(suite.T(), err)
}

func (suite *ReportFileRepositoryTestSuite) TestReportFile_GetById_Ok() {
	report := &proto.MgoReportFile{Id: bson.NewObjectId(), MerchantId: bson.NewObjectId()}
	assert.NoError(suite.T(), suite.service.Insert(report))

	rep, err := suite.service.GetById(report.Id.Hex())
	assert.NoError(suite.T(), err, "unable to get the report file")
	assert.Equal(suite.T(), report.Id, rep.Id)
	assert.Equal(suite.T(), report.MerchantId, rep.MerchantId)
}
