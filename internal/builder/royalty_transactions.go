package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"math"
)

type RoyaltyTransactions DefaultHandler

func newRoyaltyTransactionsHandler(h *Handler) BuildInterface {
	return &RoyaltyTransactions{Handler: h}
}

func (h *RoyaltyTransactions) Validate() error {
	params, err := h.GetParams()

	if err != nil {
		return err
	}

	if _, ok := params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *RoyaltyTransactions) Build() (interface{}, error) {
	params, _ := h.GetParams()
	royalty, err := h.royaltyRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	merchant, err := h.merchantRepository.GetById(royalty.MerchantId.Hex())

	if err != nil {
		return nil, err
	}

	orders, err := h.transactionsRepository.GetByRoyalty(royalty)

	if err != nil {
		return nil, err
	}

	var transactions []map[string]interface{}

	for _, order := range orders {
		netRevenue := float64(0)
		if order.NetRevenue != nil {
			netRevenue = order.NetRevenue.Amount
		}

		transactions = append(transactions, map[string]interface{}{
			"status":     order.Status,
			"project":    order.Project.Name[0].Value,
			"datetime":   order.TransactionDate.Format("2006-01-02T15:04:05"),
			"country":    order.CountryCode,
			"method":     order.PaymentMethod.Name,
			"id":         order.Id.Hex(),
			"net_amount": math.Round(netRevenue*100) / 100,
		})
	}

	result := map[string]interface{}{
		"id":                       royalty.Id.Hex(),
		"report_date":              royalty.CreatedAt.Format("2006-01-02T15:04:05"),
		"merchant_legal_name":      merchant.Company.Name,
		"merchant_company_address": merchant.Company.Address,
		"start_date":               royalty.PeriodFrom.Format("2006-01-02T15:04:05"),
		"end_date":                 royalty.PeriodTo.Format("2006-01-02T15:04:05"),
		"currency":                 royalty.Currency,
		"transactions":             transactions,
	}

	return result, nil
}

func (h *RoyaltyTransactions) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
