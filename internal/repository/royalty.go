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
	collectionRoyalty = "royalty_report"
)

type RoyaltyRepositoryInterface interface {
	GetById(string) (*billingProto.MgoRoyaltyReport, error)
}

func NewRoyaltyReportRepository(db database.SourceInterface) RoyaltyRepositoryInterface {
	s := &RoyaltyRepository{Repository: &Repository{db: db}}
	return s
}

func (h *RoyaltyRepository) GetById(id string) (*billingProto.MgoRoyaltyReport, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyalty),
			zap.String(pkg.ErrorDatabaseFieldObjectId, id),
		)
		return nil, err
	}

	report := new(billingProto.MgoRoyaltyReport)
	filter := bson.M{"_id": oid}
	err = h.db.Collection(collectionRoyalty).FindOne(h.getContext(), filter).Decode(&report)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionRoyalty),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	return report, err
}
