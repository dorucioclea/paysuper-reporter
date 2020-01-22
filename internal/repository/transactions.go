package repository

import (
	"github.com/jinzhu/now"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	database "gopkg.in/paysuper/paysuper-database-mongo.v2"
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

func NewTransactionsRepository(db database.SourceInterface) TransactionsRepositoryInterface {
	s := &TransactionsRepository{Repository: &Repository{db: db}}
	return s
}

func (h *TransactionsRepository) GetByRoyalty(report *billingProto.MgoRoyaltyReport) ([]*billingProto.MgoOrderViewPublic, error) {
	filter := bson.M{
		"merchant_id":         report.MerchantId,
		"pm_order_close_date": bson.M{"$gte": report.PeriodFrom, "$lte": report.PeriodTo},
		"status": bson.M{"$in": []string{
			constant.OrderPublicStatusProcessed,
			constant.OrderPublicStatusRefunded,
			constant.OrderPublicStatusChargeback,
		}},
	}
	opts := options.Find().SetSort(bson.M{"created_at": 1})
	cursor, err := h.db.Collection(collectionOrderView).Find(h.getContext(), filter, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionOrderView),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	result := make([]*billingProto.MgoOrderViewPublic, 0)
	err = cursor.All(h.getContext(), &result)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionOrderView),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	return result, nil
}

func (h *TransactionsRepository) GetByVat(report *billingProto.MgoVatReport) ([]*billingProto.MgoOrderViewPrivate, error) {
	filter := bson.M{
		"pm_order_close_date": bson.M{
			"$gte": now.New(report.DateFrom).BeginningOfDay(),
			"$lte": now.New(report.DateTo).EndOfDay(),
		},
		"country_code": report.Country,
	}
	opts := options.Find().SetSort(bson.M{"created_at": 1})
	cursor, err := h.db.Collection(collectionOrderView).Find(h.getContext(), filter, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionOrderView),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	result := make([]*billingProto.MgoOrderViewPrivate, 0)
	err = cursor.All(h.getContext(), &result)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionOrderView),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
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
	oid, err := primitive.ObjectIDFromHex(merchantId)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionVat),
			zap.String("merchant_id", merchantId),
		)
		return nil, err
	}

	filter := make(bson.M)
	filter["merchant_id"] = oid

	if len(paymentMethods) > 0 {
		var paymentMethod []primitive.ObjectID

		for _, v := range paymentMethods {
			oid, err = primitive.ObjectIDFromHex(v)

			if err != nil {
				zap.L().Error(
					pkg.ErrorDatabaseInvalidObjectId,
					zap.Error(err),
					zap.String(pkg.ErrorDatabaseFieldCollection, collectionVat),
					zap.String("payment_method_id", v),
				)
				continue
			}

			paymentMethod = append(paymentMethod, oid)
		}

		filter["payment_method._id"] = bson.M{"$in": paymentMethod}
	}

	if len(status) > 0 {
		filter["status"] = bson.M{"$in": status}
	}

	pmDates := make(bson.M)

	if dateFrom != 0 {
		pmDates["$gte"] = time.Unix(dateFrom, 0)
	}

	if dateTo != 0 {
		pmDates["$lte"] = time.Unix(dateTo, 0)
	}

	if len(pmDates) > 0 {
		filter["pm_order_close_date"] = pmDates
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1})
	cursor, err := h.db.Collection(collectionOrderView).Find(h.getContext(), filter, opts)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionOrderView),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	result := make([]*billingProto.MgoOrderViewPublic, 0)
	err = cursor.All(h.getContext(), &result)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionOrderView),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	return result, nil
}
