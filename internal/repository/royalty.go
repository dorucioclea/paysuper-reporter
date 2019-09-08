package repository

import (
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
)

const (
	collectionRoyalty = "royalty_report"
)

type RoyaltyRepositoryInterface interface {
	Insert(*billingProto.MgoRoyaltyReport) error
	GetById(string) (*billingProto.MgoRoyaltyReport, error)
}

func NewRoyaltyReportRepository(db *database.Source) RoyaltyRepositoryInterface {
	s := &RoyaltyRepository{db: db}
	return s
}

func (h *RoyaltyRepository) Insert(report *billingProto.MgoRoyaltyReport) error {
	return h.db.Collection(collectionRoyalty).Insert(report)
}

func (h *RoyaltyRepository) GetById(id string) (*billingProto.MgoRoyaltyReport, error) {
	var report *billingProto.MgoRoyaltyReport
	err := h.db.Collection(collectionRoyalty).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionRoyalty),
			zap.String("id", id),
		)
	}

	return report, err
}
