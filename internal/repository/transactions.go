package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/jinzhu/now"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const (
	collectionOrderView = "order_view"
)

type TransactionsRepositoryInterface interface {
	GetByRoyalty(*billingProto.MgoRoyaltyReport) ([]*billingProto.MgoOrderViewPublic, error)
	GetByVat(*billingProto.MgoVatReport) ([]*billingProto.MgoOrderViewPrivate, error)
	FindByMerchant(string, []string, []string, int64, int64) ([]*billingProto.MgoOrderViewPublic, error)
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
		"status": bson.M{"$in": []string{
			constant.OrderPublicStatusProcessed,
			constant.OrderPublicStatusRefunded,
			constant.OrderPublicStatusChargeback,
		}},
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

func (h *TransactionsRepository) FindByMerchant(
	merchantId string,
	status []string,
	paymentMethods []string,
	dateFrom int64,
	dateTo int64,
) ([]*billingProto.MgoOrderViewPublic, error) {
	var result []*billingProto.MgoOrderViewPublic

	query := make(bson.M)
	query["merchant_id"] = bson.ObjectIdHex(merchantId)

	if len(paymentMethods) > 0 {
		var paymentMethod []bson.ObjectId

		for _, v := range paymentMethods {
			paymentMethod = append(paymentMethod, bson.ObjectIdHex(v))
		}

		query["payment_method._id"] = bson.M{"$in": paymentMethod}
	}

	if len(status) > 0 {
		query["status"] = bson.M{"$in": status}
	}

	pmDates := make(bson.M)

	if dateFrom != 0 {
		pmDates["$gte"] = time.Unix(dateFrom, 0)
	}

	if dateTo != 0 {
		pmDates["$lte"] = time.Unix(dateTo, 0)
	}

	if len(pmDates) > 0 {
		query["pm_order_close_date"] = pmDates
	}
	zap.L().Error("transaction search", zap.Any("query", query))
	err := h.db.Collection(collectionOrderView).Find(query).Sort("-created_at").All(&result)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionOrderView),
			zap.Any("match", nil),
		)
		return nil, err
	}

	return result, nil
}
