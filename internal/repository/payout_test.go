package repository

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"testing"
)

type PayoutRepositoryTestSuite struct {
	suite.Suite
	db      mongodb.SourceInterface
	service PayoutRepositoryInterface
	log     *zap.Logger
}

func Test_PayoutRepository(t *testing.T) {
	suite.Run(t, new(PayoutRepositoryTestSuite))
}

func (suite *PayoutRepositoryTestSuite) SetupTest() {
	cfg, err := config.NewConfig()
	if err != nil {
		suite.FailNow("Config load failed", "%v", err)
	}

	m, err := migrate.New("file://../../migrations/tests", cfg.Db.Dsn)
	assert.NoError(suite.T(), err, "Migrate init failed")

	err = m.Up()
	if err != nil && err.Error() != "no change" {
		suite.FailNow("Migrations failed", "%v", err)
	}

	suite.db, err = mongodb.NewDatabase()

	if err != nil {
		suite.FailNow("Database connection failed", "%v", err)
	}

	suite.log, err = zap.NewProduction()

	if err != nil {
		suite.FailNow("Logger initialization failed", "%v", err)
	}

	suite.service = NewPayoutRepository(suite.db)
}

func (suite *PayoutRepositoryTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	_ = suite.db.Close()
}

func (suite *PayoutRepositoryTestSuite) TestPayoutRepository_GetById_Error() {
	_, err := suite.service.GetById("ffffffffffffffffffffffff")
	assert.Error(suite.T(), err)
}

func (suite *PayoutRepositoryTestSuite) TestPayoutRepository_GetById_Ok() {
	id, err := primitive.ObjectIDFromHex("5ced34d689fce60bf4440829")
	assert.NoError(suite.T(), err)
	rep, err := suite.service.GetById(id.Hex())

	assert.NoError(suite.T(), err, "unable to get the payout report")
	assert.Equal(suite.T(), id, rep.Id)
}
