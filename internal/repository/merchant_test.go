package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
)

type MerchantRepositoryTestSuite struct {
	suite.Suite
	db      *mongodb.Source
	service MerchantRepositoryInterface
	log     *zap.Logger
}

func Test_MerchantRepository(t *testing.T) {
	suite.Run(t, new(MerchantRepositoryTestSuite))
}

func (suite *MerchantRepositoryTestSuite) SetupTest() {
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

	suite.service = NewMerchantRepository(suite.db)
}

func (suite *MerchantRepositoryTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	suite.db.Close()
}

func (suite *MerchantRepositoryTestSuite) TestMerchantRepository_GetById_Error() {
	_, err := suite.service.GetById(bson.NewObjectId().Hex())
	assert.Error(suite.T(), err)
}

func (suite *MerchantRepositoryTestSuite) TestMerchantRepository_GetById_Ok() {
	id := bson.ObjectIdHex("5ced34d689fce60bf4440829")
	rep, err := suite.service.GetById(id.Hex())

	assert.NoError(suite.T(), err, "unable to get the merchant")
	assert.Equal(suite.T(), id, rep.Id)
}