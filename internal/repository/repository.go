package repository

import (
	database "gopkg.in/paysuper/paysuper-database-mongo.v2"
)

type Repository struct {
	db database.SourceInterface
}

type RoyaltyRepository Repository
type VatRepository Repository
type TransactionsRepository Repository
type PayoutRepository Repository
type MerchantRepository Repository
