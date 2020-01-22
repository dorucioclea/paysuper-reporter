package repository

import (
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-reporter/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	database "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionMerchant = "merchant"
)

type MerchantRepositoryInterface interface {
	GetById(string) (*billingProto.MgoMerchant, error)
}

func NewMerchantRepository(db database.SourceInterface) MerchantRepositoryInterface {
	s := &MerchantRepository{Repository: &Repository{db: db}}
	return s
}

func (h *MerchantRepository) GetById(id string) (*billingProto.MgoMerchant, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionMerchant),
			zap.String(pkg.ErrorDatabaseFieldObjectId, id),
		)
		return nil, err
	}

	merchant := new(billingProto.MgoMerchant)
	filter := bson.M{"_id": oid}
	err = h.db.Collection(collectionMerchant).FindOne(h.getContext(), filter).Decode(&merchant)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayout),
			zap.String(pkg.ErrorDatabaseFieldObjectId, id),
		)
		return nil, err
	}

	return merchant, err
}
