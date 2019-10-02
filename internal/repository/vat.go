package repository

import (
	"github.com/globalsign/mgo/bson"
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
	GetByCountry(string) ([]*billingProto.MgoVatReport, error)
}

func NewVatRepository(db *database.Source) VatRepositoryInterface {
	s := &VatRepository{db: db}
	return s
}

func (h *VatRepository) GetById(id string) (*billingProto.MgoVatReport, error) {
	var report *billingProto.MgoVatReport

	err := h.db.Collection(collectionVat).FindId(bson.ObjectIdHex(id)).One(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionVat),
			zap.String("id", id),
		)
	}

	return report, err
}

func (h *VatRepository) GetByCountry(country string) ([]*billingProto.MgoVatReport, error) {
	var report []*billingProto.MgoVatReport

	err := h.db.Collection(collectionVat).Find(bson.M{"country": country}).All(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionVat),
			zap.String("country", country),
		)
	}

	return report, err
}
