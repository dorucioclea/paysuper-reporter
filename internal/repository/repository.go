package repository

import (
	"context"
	database "gopkg.in/paysuper/paysuper-database-mongo.v2"
	"time"
)

type Repository struct {
	db database.SourceInterface
}

type RoyaltyRepository struct {
	*Repository
}
type VatRepository RoyaltyRepository
type TransactionsRepository RoyaltyRepository
type PayoutRepository RoyaltyRepository
type MerchantRepository RoyaltyRepository

func (m *Repository) getContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return ctx
}
