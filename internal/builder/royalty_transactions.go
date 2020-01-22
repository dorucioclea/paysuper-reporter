package builder

import (
	"context"
	"errors"
	"fmt"
	billingGrpc "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
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

	_, err = primitive.ObjectIDFromHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
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

	res, err := h.billing.GetOperatingCompany(
		context.Background(),
		&billingGrpc.GetOperatingCompanyRequest{Id: royalty.OperatingCompanyId},
	)

	if err != nil || res.Company == nil {
		if err == nil {
			err = errors.New(res.Message.Message)
		}

		zap.L().Error(
			"unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", royalty.OperatingCompanyId),
		)

		return nil, err
	}

	result := map[string]interface{}{
		"id":                       royalty.Id.Hex(),
		"report_date":              royalty.CreatedAt.Format("2006-01-02"),
		"merchant_legal_name":      merchant.Company.Name,
		"merchant_company_address": merchant.Company.Address,
		"start_date":               royalty.PeriodFrom.Format("2006-01-02"),
		"end_date":                 royalty.PeriodTo.Format("2006-01-02"),
		"currency":                 royalty.Currency,
		"oc_name":                  res.Company.Name,
		"oc_address":               res.Company.Address,
		"transactions":             transactions,
	}

	return result, nil
}

func (h *RoyaltyTransactions) PostProcess(_ context.Context, _, _ string, _ int64, _ []byte) error {
	return nil
}
