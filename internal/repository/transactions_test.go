package repository

import (
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
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
	var err error

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

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_Insert_Error() {
	report := &billingProto.MgoOrderViewPublic{}
	err := suite.service.Insert(report)
	assert.Error(suite.T(), err)
}

func (suite *TransactionsRepositoryTestSuite) TestTransactionsRepository_Insert_Ok() {
	report := &billingProto.MgoOrderViewPublic{Id: bson.NewObjectId(), MerchantId: bson.NewObjectId()}
	err := suite.service.Insert(report)
	assert.NoError(suite.T(), err, "unable to insert the transaction")
}

func (suite *TransactionsRepositoryTestSuite) TestVatRepository_GetByRoyalty_Ok() {
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		MerchantId: bson.NewObjectId(),
		PeriodFrom: time.Now().AddDate(0, 0, -1),
		PeriodTo:   time.Now().AddDate(0, 0, 1),
	}
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		MerchantId:      report.MerchantId,
		TransactionDate: time.Now(),
		Status:          constant.OrderPublicStatusProcessed,
	}
	assert.NoError(suite.T(), suite.service.Insert(order))

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 1)
}

func (suite *TransactionsRepositoryTestSuite) TestVatRepository_GetByRoyalty_Error_RangeDates() {
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		MerchantId: bson.NewObjectId(),
		PeriodFrom: time.Now().AddDate(0, -1, -1),
		PeriodTo:   time.Now().AddDate(0, 0, -1),
	}
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		MerchantId:      report.MerchantId,
		TransactionDate: time.Now(),
		Status:          constant.OrderPublicStatusProcessed,
	}
	assert.NoError(suite.T(), suite.service.Insert(order))

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestVatRepository_GetByRoyalty_Error_UnexistsStatus() {
	report := &billingProto.MgoRoyaltyReport{
		Id:         bson.NewObjectId(),
		MerchantId: bson.NewObjectId(),
		PeriodFrom: time.Now().AddDate(0, 0, -1),
		PeriodTo:   time.Now().AddDate(0, 0, 1),
	}
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		MerchantId:      report.MerchantId,
		TransactionDate: time.Now(),
		Status:          constant.OrderPublicStatusCreated,
	}
	assert.NoError(suite.T(), suite.service.Insert(order))

	orders, err := suite.service.GetByRoyalty(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestVatRepository_GetByVat_Ok() {
	report := &billingProto.MgoVatReport{
		Id:       bson.NewObjectId(),
		DateFrom: time.Now(),
		DateTo:   time.Now(),
		Country:  "RU",
	}
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		TransactionDate: time.Now(),
		CountryCode:     report.Country,
	}
	assert.NoError(suite.T(), suite.service.Insert(order))

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 1)
}

func (suite *TransactionsRepositoryTestSuite) TestVatRepository_GetByVat_Error_RangeDate() {
	report := &billingProto.MgoVatReport{
		Id:       bson.NewObjectId(),
		DateFrom: time.Now().AddDate(0, 0, -1),
		DateTo:   time.Now().AddDate(0, 0, 1),
		Country:  "RU",
	}
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		TransactionDate: time.Now(),
		CountryCode:     report.Country,
	}
	assert.NoError(suite.T(), suite.service.Insert(order))

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}

func (suite *TransactionsRepositoryTestSuite) TestVatRepository_GetByVat_Error_Country() {
	report := &billingProto.MgoVatReport{
		Id:       bson.NewObjectId(),
		DateFrom: time.Now(),
		DateTo:   time.Now(),
		Country:  "RU",
	}
	order := &billingProto.MgoOrderViewPublic{
		Id:              bson.NewObjectId(),
		TransactionDate: time.Now(),
		CountryCode:     "UA",
	}
	assert.NoError(suite.T(), suite.service.Insert(order))

	orders, err := suite.service.GetByVat(report)
	assert.NoError(suite.T(), err, "unable to get the orders")
	assert.Len(suite.T(), orders, 0)
}
