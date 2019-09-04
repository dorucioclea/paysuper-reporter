package repository

import (
	database "github.com/paysuper/paysuper-database-mongo"
)

type Repository struct {
	db *database.Source
}

type ReportFileRepository Repository
type RoyaltyRepository Repository
type VatRepository Repository
type TransactionsRepository Repository
