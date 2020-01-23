package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
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

	if bson.IsObjectIdHex(h.report.MerchantId) != true {
		return errors.New(errs.ErrorParamMerchantIdNotFound.Message)
	}

	if _, ok := params[reporterpb.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	_, err = primitive.ObjectIDFromHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *RoyaltyTransactions) Build() (interface{}, error) {
	ctx := context.TODO()
	params, _ := h.GetParams()
	royaltyId := fmt.Sprintf("%s", params[reporterpb.ParamsFieldId])

	royaltyRequest := &billingpb.GetRoyaltyReportRequest{ReportId: royaltyId, MerchantId: h.report.MerchantId}
	royalty, err := h.billing.GetRoyaltyReport(ctx, royaltyRequest)

	if err != nil || royalty.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(royalty.Message.Message)
		}

		zap.L().Error(
			"Unable to get royalty report",
			zap.Error(err),
			zap.String("royalty_id", royaltyId),
		)

		return nil, err
	}

	merchantRequest := &billingpb.GetMerchantByRequest{MerchantId: h.report.MerchantId}
	merchant, err := h.billing.GetMerchantBy(ctx, merchantRequest)

	if err != nil || merchant.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(merchant.Message.Message)
		}

		zap.L().Error(
			"Unable to get merchant",
			zap.Error(err),
			zap.String("merchant_id", h.report.MerchantId),
		)

		return nil, err
	}

	ordersRequest := &billingpb.ListOrdersRequest{Merchant: []string{h.report.MerchantId}}
	orders, err := h.billing.FindAllOrdersPublic(ctx, ordersRequest)

	if err != nil || orders.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(orders.Message.Message)
		}

		zap.L().Error(
			"Unable to get orders",
			zap.Error(err),
			zap.String("merchant_id", h.report.MerchantId),
		)

		return nil, err
	}

	var transactions []map[string]interface{}

	for _, order := range orders.Item.Items {
		netRevenue := float64(0)
		if order.NetRevenue != nil {
			netRevenue = order.NetRevenue.Amount
		}

		datetime, err := ptypes.Timestamp(order.TransactionDate)

		if err != nil {
			zap.L().Error(
				"Unable to cast timestamp to time",
				zap.Error(err),
				zap.String("transaction_date", order.TransactionDate.String()),
			)
			return nil, err
		}

		transactions = append(transactions, map[string]interface{}{
			"status":     order.Status,
			"project":    order.Project.Name["en"],
			"datetime":   datetime.Format("2006-01-02T15:04:05"),
			"country":    order.CountryCode,
			"method":     order.PaymentMethod.Name,
			"id":         order.Id,
			"net_amount": math.Round(netRevenue*100) / 100,
		})
	}

	ocRequest := &billingpb.GetOperatingCompanyRequest{Id: royalty.Item.OperatingCompanyId}
	operatingCompany, err := h.billing.GetOperatingCompany(ctx, ocRequest)

	if err != nil || operatingCompany.Company == nil {
		if err == nil {
			err = errors.New(operatingCompany.Message.Message)
		}

		zap.L().Error(
			"Unable to get operating company",
			zap.Error(err),
			zap.String("operating_company_id", royalty.Item.OperatingCompanyId),
		)

		return nil, err
	}

	date, err := ptypes.Timestamp(royalty.Item.CreatedAt)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("created_at", royalty.Item.CreatedAt.String()),
		)
		return nil, err
	}

	periodFrom, err := ptypes.Timestamp(royalty.Item.PeriodFrom)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("period_from", royalty.Item.PeriodFrom.String()),
		)
		return nil, err
	}

	periodTo, err := ptypes.Timestamp(royalty.Item.PeriodTo)

	if err != nil {
		zap.L().Error(
			"Unable to cast timestamp to time",
			zap.Error(err),
			zap.String("period_to", royalty.Item.PeriodTo.String()),
		)
		return nil, err
	}

	result := map[string]interface{}{
		"id":                       royalty.Item.Id,
		"report_date":              date.Format("2006-01-02"),
		"merchant_legal_name":      merchant.Item.Company.Name,
		"merchant_company_address": merchant.Item.Company.Address,
		"start_date":               periodFrom.Format("2006-01-02"),
		"end_date":                 periodTo.Format("2006-01-02"),
		"currency":                 royalty.Item.Currency,
		"oc_name":                  operatingCompany.Company.Name,
		"oc_address":               operatingCompany.Company.Address,
		"transactions":             transactions,
	}

	return result, nil
}

func (h *RoyaltyTransactions) PostProcess(_ context.Context, _, _ string, _ int64, _ []byte) error {
	return nil
}
