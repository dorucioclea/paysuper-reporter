package repository

import (
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	mongodb "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"testing"
	"time"
)

type TransactionsRepositoryTestSuite struct {
	suite.Suite
	db      mongodb.SourceInterface
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

	_ = suite.db.Close()
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByRoyalty_Ok() {
	oid, err := primitive.ObjectIDFromHex("5ced34d689fce60bf444082a")
	assert.NoError(suite.T(), err)
	report := &billingProto.MgoRoyaltyReport{
		MerchantId: oid,
		PeriodFrom: time.Unix(1562258329, 0).AddDate(0, 0, -1),
		PeriodTo:   time.Unix(1562258329, 0).AddDate(0, 0, 1),
	}

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 1)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByRoyalty_Error_RangeDates() {
	order := &billingProto.MgoOrderViewPublic{
		Id:              primitive.NewObjectID(),
		MerchantId:      primitive.NewObjectID(),
		TransactionDate: time.Unix(1562258329, 0),
		Status:          constant.OrderPublicStatusProcessed,
	}
	report := &billingProto.MgoRoyaltyReport{
		Id:         primitive.NewObjectID(),
		MerchantId: order.MerchantId,
		PeriodFrom: time.Now().AddDate(0, -1, -1),
		PeriodTo:   time.Now().AddDate(0, 0, -1),
	}

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByRoyalty_Error_UnexistsStatus() {
	order := &billingProto.MgoOrderViewPrivate{
		Id:              primitive.NewObjectID(),
		MerchantId:      primitive.NewObjectID(),
		TransactionDate: time.Unix(1562258329, 0),
		Status:          constant.OrderPublicStatusCreated,
	}
	report := &billingProto.MgoRoyaltyReport{
		Id:         primitive.NewObjectID(),
		MerchantId: order.MerchantId,
		PeriodFrom: time.Now().AddDate(0, 0, -1),
		PeriodTo:   time.Now().AddDate(0, 0, 1),
	}

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_GetByVat_Ok() {
	oid, err := primitive.ObjectIDFromHex("5ced34d689fce60bf4440829")
	assert.NoError(suite.T(), err)
	report := &billingProto.MgoVatReport{
		Id:       oid,
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
		Id:       primitive.NewObjectID(),
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
		Id:       primitive.NewObjectID(),
		DateFrom: time.Now(),
		DateTo:   time.Now(),
		Country:  "RU",
	}

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}
