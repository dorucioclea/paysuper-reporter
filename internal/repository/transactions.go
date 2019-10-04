package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/jinzhu/now"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
)

const (
	collectionOrderView = "order_view"
)

type TransactionsRepositoryInterface interface {
	GetByRoyalty(*billingProto.MgoRoyaltyReport) ([]*billingProto.MgoOrderViewPublic, error)
	GetByVat(*billingProto.MgoVatReport) ([]*billingProto.MgoOrderViewPrivate, error)
}

func NewTransactionsRepository(db *database.Source) TransactionsRepositoryInterface {
	s := &TransactionsRepository{db: db}
	return s
}

func (h *TransactionsRepository) GetByRoyalty(report *billingProto.MgoRoyaltyReport) ([]*billingProto.MgoOrderViewPublic, error) {
	var result []*billingProto.MgoOrderViewPublic

	match := bson.M{
		"merchant_id":         report.MerchantId,
		"pm_order_close_date": bson.M{"$gte": report.PeriodFrom, "$lte": report.PeriodTo},
		"status":              constant.OrderPublicStatusProcessed,
	}
	err := h.db.Collection(collectionOrderView).Find(match).Sort("created_at").All(&result)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionOrderView),
			zap.Any("match", match),
		)
		return nil, err
	}

	return result, nil
}

func (h *TransactionsRepository) GetByVat(report *billingProto.MgoVatReport) ([]*billingProto.MgoOrderViewPrivate, error) {
	var result []*billingProto.MgoOrderViewPrivate

	match := bson.M{
		"pm_order_close_date": bson.M{
			"$gte": now.New(report.DateFrom).BeginningOfDay(),
			"$lte": now.New(report.DateTo).EndOfDay(),
		},
		"country_code": report.Country,
	}
	err := h.db.Collection(collectionOrderView).Find(match).Sort("created_at").All(&result)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionOrderView),
			zap.Any("match", match),
		)
		return nil, err
	}

	return result, nil
}
