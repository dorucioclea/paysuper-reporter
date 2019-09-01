package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
)

const (
	collectionRoyaltyReport = "royalty_report"
	collectionOrderView     = "order_view"
)

type RoyaltyReportRepositoryInterface interface {
	GetById(string) (*proto.RoyaltyReport, error)
	GetTransactions(*proto.RoyaltyReport) ([]*proto.OrderViewPublic, error)
}

func NewRoyaltyReportRepository(db *database.Source) RoyaltyReportRepositoryInterface {
	s := &RoyaltyReportRepository{db: db}
	return s
}

func (h *RoyaltyReportRepository) GetById(id string) (*proto.RoyaltyReport, error) {
	var report *proto.RoyaltyReport
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

func (h *RoyaltyReportRepository) GetTransactions(report *proto.RoyaltyReport) ([]*proto.OrderViewPublic, error) {
	var result []*proto.OrderViewPublic

	from, _ := ptypes.Timestamp(report.PeriodFrom)
	to, _ := ptypes.Timestamp(report.PeriodTo)

	match := bson.M{
		"merchant_id":         bson.ObjectIdHex(report.MerchantId),
		"pm_order_close_date": bson.M{"$gte": from, "$lte": to},
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
