package repository

import (
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
)

const (
	collectionMerchant = "merchant"
)

type MerchantRepositoryInterface interface {
	GetById(string) (*billingProto.MgoMerchant, error)
}

func NewMerchantRepository(db *database.Source) MerchantRepositoryInterface {
	s := &MerchantRepository{db: db}
	return s
}

func (h *MerchantRepository) GetById(id string) (*billingProto.MgoMerchant, error) {
	var merchant *billingProto.MgoMerchant

	err := h.db.Collection(collectionMerchant).FindId(bson.ObjectIdHex(id)).One(&merchant)

	if err != nil {
		zap.L().Error(
			errors.ErrorDatabaseQueryFailed.Message,
			zap.Error(err),
			zap.String("collection", collectionPayout),
			zap.String("id", id),
		)
	}

	return merchant, err
}
