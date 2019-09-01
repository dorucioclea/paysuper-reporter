package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/internal/helpers"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
)

const (
	collectionVatReports = "vat_reports"
)

type VatReportRepositoryInterface interface {
	GetById(string) (*proto.VatReport, error)
	GetTransactions(*proto.VatReport) ([]*proto.OrderViewPublic, error)
}

func NewVatReportRepository(db *database.Source) VatReportRepositoryInterface {
	s := &VatReportRepository{db: db}
	return s
}

func (h *VatReportRepository) GetById(id string) (*proto.VatReport, error) {
	var report *proto.VatReport

	query := bson.M{
		"status": bson.M{
			"$in": []string{
				billingPkg.VatReportStatusThreshold,
				billingPkg.VatReportStatusNeedToPay,
				billingPkg.VatReportStatusOverdue,
			},
		},
	}
	err := h.db.Collection(collectionVatReports).Find(query).One(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionRoyaltyReport),
			zap.String("id", id),
		)
	}

	return nil, err
}

func (h *VatReportRepository) GetTransactions(report *proto.VatReport) ([]*proto.OrderViewPublic, error) {
	var result []*proto.OrderViewPublic

	from, _ := ptypes.Timestamp(report.DateFrom)
	to, _ := ptypes.Timestamp(report.DateTo)

	match := bson.M{
		"pm_order_close_date": bson.M{
			"$gte": helpers.BeginOfDay(from),
			"$lte": helpers.EndOfDay(to),
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
