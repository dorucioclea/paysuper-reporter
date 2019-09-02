package repository

import (
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
)

const (
	collectionRoyaltyReport = "royalty_report"
	collectionOrderView     = "order_view"
)

type RoyaltyReportRepositoryInterface interface {
	GetById(string) (*billingProto.MgoRoyaltyReport, error)
	GetTransactions(*billingProto.MgoRoyaltyReport) ([]*billingProto.MgoOrderViewPublic, error)
}

func NewRoyaltyReportRepository(db *database.Source) RoyaltyReportRepositoryInterface {
	s := &RoyaltyReportRepository{db: db}
	return s
}

func (h *RoyaltyReportRepository) GetById(id string) (*billingProto.MgoRoyaltyReport, error) {
	var report *billingProto.MgoRoyaltyReport
	err := h.db.Collection(collectionRoyaltyReport).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionRoyaltyReport),
			zap.String("id", id),
		)
	}

	return report, err
}

func (h *RoyaltyReportRepository) GetTransactions(report *billingProto.MgoRoyaltyReport) ([]*billingProto.MgoOrderViewPublic, error) {
	var result []*billingProto.MgoOrderViewPublic

	match := bson.M{
		"merchant_id":         report.MerchantId.Hex(),
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
