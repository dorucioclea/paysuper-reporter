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
	collectionPayout = "payout_documents"
)

type PayoutRepositoryInterface interface {
	GetById(string) (*billingProto.MgoPayoutDocument, error)
}

func NewPayoutRepository(db database.SourceInterface) PayoutRepositoryInterface {
	s := &PayoutRepository{Repository: &Repository{db: db}}
	return s
}

func (h *PayoutRepository) GetById(id string) (*billingProto.MgoPayoutDocument, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayout),
			zap.String(pkg.ErrorDatabaseFieldObjectId, id),
		)
		return nil, err
	}

	report := new(billingProto.MgoPayoutDocument)
	filter := bson.M{"_id": oid}
	err = h.db.Collection(collectionPayout).FindOne(h.getContext(), filter).Decode(&report)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionPayout),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	return report, err
}
