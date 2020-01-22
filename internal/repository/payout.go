package repository

import (
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
	database "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionPayout = "payout_documents"
)

type PayoutRepositoryInterface interface {
	GetById(string) (*billingProto.MgoPayoutDocument, error)
}

func NewPayoutRepository(db *database.Source) PayoutRepositoryInterface {
	s := &PayoutRepository{db: db}
	return s
}

func (h *PayoutRepository) GetById(id string) (*billingProto.MgoPayoutDocument, error) {
	var report *billingProto.MgoPayoutDocument

	err := h.db.Collection(collectionPayout).FindId(bson.ObjectIdHex(id)).One(&report)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionPayout),
			zap.String("id", id),
		)
	}

	return report, err
}
