package repository

import (
	"github.com/globalsign/mgo/bson"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/internal/helpers"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
)

const (
	collectionVatReports = "vat_reports"
)

type VatReportRepositoryInterface interface {
	GetById(string) (*billingProto.MgoVatReport, error)
	GetTransactions(*billingProto.MgoVatReport) ([]*billingProto.MgoOrderViewPublic, error)
}

func NewVatReportRepository(db *database.Source) VatReportRepositoryInterface {
	s := &VatReportRepository{db: db}
	return s
}

func (h *VatReportRepository) GetById(id string) (*billingProto.MgoVatReport, error) {
	var report *billingProto.MgoVatReport

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

func (h *VatReportRepository) GetTransactions(report *billingProto.MgoVatReport) ([]*billingProto.MgoOrderViewPublic, error) {
	var result []*billingProto.MgoOrderViewPublic

	match := bson.M{
		"pm_order_close_date": bson.M{
			"$gte": helpers.BeginOfDay(report.DateFrom),
			"$lte": helpers.EndOfDay(report.DateTo),
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
