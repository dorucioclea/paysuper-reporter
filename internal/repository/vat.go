package repository

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-reporter/pkg"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	database "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

const (
	collectionVat = "vat_reports"
)

type VatRepositoryInterface interface {
	GetById(string) (*billingProto.MgoVatReport, error)
	GetByCountry(string) ([]*billingProto.MgoVatReport, error)
}

func NewVatRepository(db database.SourceInterface) VatRepositoryInterface {
	s := &VatRepository{db: db}
	return s
}

func (h *VatRepository) GetById(id string) (*billingProto.MgoVatReport, error) {
	oid, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseInvalidObjectId,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionVat),
			zap.String(pkg.ErrorDatabaseFieldObjectId, id),
		)
		return nil, err
	}

	report := new(billingProto.MgoVatReport)
	filter := bson.M{"_id": oid}
	err = h.db.Collection(collectionVat).FindOne(context.Background(), filter).Decode(&report)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionVat),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
	}

	return report, err
}

func (h *VatRepository) GetByCountry(country string) ([]*billingProto.MgoVatReport, error) {
	filter := bson.M{"country": country}
	cursor, err := h.db.Collection(collectionVat).Find(context.Background(), filter)

	if err != nil {
		zap.L().Error(
			pkg.ErrorDatabaseQueryFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionVat),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	report := make([]*billingProto.MgoVatReport, 0)
	err = cursor.All(context.Background(), &report)

	if err != nil {
		zap.L().Error(
			pkg.ErrorQueryCursorExecutionFailed,
			zap.Error(err),
			zap.String(pkg.ErrorDatabaseFieldCollection, collectionVat),
			zap.Any(pkg.ErrorDatabaseFieldQuery, filter),
		)
		return nil, err
	}

	return report, err
}
