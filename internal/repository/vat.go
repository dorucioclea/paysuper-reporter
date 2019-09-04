package repository

import (
	"github.com/globalsign/mgo/bson"
	billingPkg "github.com/paysuper/paysuper-billing-server/pkg"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
)

const (
	collectionVat = "vat_reports"
)

type VatRepositoryInterface interface {
	GetById(string) (*billingProto.MgoVatReport, error)
}

func NewVatRepository(db *database.Source) VatRepositoryInterface {
	s := &VatRepository{db: db}
	return s
}

func (h *VatRepository) GetById(id string) (*billingProto.MgoVatReport, error) {
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
	err := h.db.Collection(collectionVat).Find(query).One(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionRoyalty),
			zap.String("id", id),
		)
	}

	return nil, err
}
