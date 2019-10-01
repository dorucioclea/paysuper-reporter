package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"time"
)

type TransactionsRepositoryTestSuite struct {
	suite.Suite
	db      *mongodb.Source
	service TransactionsRepositoryInterface
	log     *zap.Logger
}

func Test_TransactionsRepository(t *testing.T) {
	suite.Run(t, new(TransactionsRepositoryTestSuite))
}

func (suite *TransactionsRepositoryTestSuite) SetupTest() {
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

	suite.service = NewTransactionsRepository(suite.db)
}

func (suite *TransactionsRepositoryTestSuite) TearDownTest() {
	if err := suite.db.Drop(); err != nil {
		suite.FailNow("Database deletion failed", "%v", err)
	}

	suite.db.Close()
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByRoyalty_Ok() {
	report := &billingProto.MgoRoyaltyReport{
		MerchantId: bson.ObjectIdHex("5ced34d689fce60bf444082a"),
		PeriodFrom: time.Unix(1562258329, 0).AddDate(0, 0, -1),
		PeriodTo:   time.Unix(1562258329, 0).AddDate(0, 0, 1),
	}

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 1)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByRoyalty_Error_RangeDates() {
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		MerchantId:      bson.NewObjectId(),
		TransactionDate: time.Unix(1562258329, 0),
		Status:          constant.OrderPublicStatusProcessed,
	}
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		MerchantId: order.MerchantId,
		PeriodFrom: time.Now().AddDate(0, -1, -1),
		PeriodTo:   time.Now().AddDate(0, 0, -1),
	}

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByRoyalty_Error_UnexistsStatus() {
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		MerchantId:      bson.NewObjectId(),
		TransactionDate: time.Unix(1562258329, 0),
		Status:          constant.OrderPublicStatusCreated,
	}
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		MerchantId: order.MerchantId,
		PeriodFrom: time.Now().AddDate(0, 0, -1),
		PeriodTo:   time.Now().AddDate(0, 0, 1),
	}

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByVat_Ok() {
	report := &billingProto.MgoVatReport{
		Id:       bson.ObjectIdHex("5ced34d689fce60bf4440829"),
		DateFrom: time.Unix(1562258329, 0),
		DateTo:   time.Unix(1562258329, 0),
		Country:  "RU",
	}

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 1)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByVat_Error_RangeDate() {
	report := &billingProto.MgoVatReport{
		Id:       bson.NewObjectId(),
		DateFrom: time.Now().AddDate(0, 0, -2),
		DateTo:   time.Now().AddDate(0, 0, -1),
		Country:  "RU",
	}

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByVat_Error_Country() {
	report := &billingProto.MgoVatReport{
		Id:       bson.NewObjectId(),
		DateFrom: time.Now(),
		DateTo:   time.Now(),
		Country:  "RU",
	}

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}
