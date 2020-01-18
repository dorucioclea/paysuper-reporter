package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.uber.org/zap"
	"math"
	"reflect"
)

type Transactions DefaultHandler

func newTransactionsHandler(h *Handler) BuildInterface {
	return &Transactions{Handler: h}
}

func (h *Transactions) Validate() error {
	params, _ := h.GetParams()

	if bson.IsObjectIdHex(h.report.MerchantId) != true {
		return errors.New(errs.ErrorParamMerchantIdNotFound.Message)
	}

	if st, ok := params[reporterpb.ParamsFieldStatus]; ok && st != nil {
		if reflect.TypeOf(st).Kind() != reflect.Slice {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}
	}

	if st, ok := params[reporterpb.ParamsFieldPaymentMethod]; ok && st != nil {
		if reflect.TypeOf(st).Kind() != reflect.Slice {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}

		for _, str := range st.([]interface{}) {
			if !bson.IsObjectIdHex(fmt.Sprintf("%s", str)) {
				return errors.New(errs.ErrorHandlerValidation.Message)
			}
		}
	}

	if st, ok := params[reporterpb.ParamsFieldDateFrom]; ok {
		if reflect.TypeOf(st).Kind() != reflect.Float64 {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}
	}

	if st, ok := params[reporterpb.ParamsFieldDateTo]; ok {
		if reflect.TypeOf(st).Kind() != reflect.Float64 {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}
	}

	return nil
}

func (h *Transactions) Build() (interface{}, error) {
	var logs []map[string]interface{}
	var status []string
	var paymentMethods []string

	dateFrom := int64(0)
	dateTo := int64(0)

	ctx := context.TODO()
	params, _ := h.GetParams()

	if st, ok := params[reporterpb.ParamsFieldStatus]; ok && st != nil {
		for _, str := range st.([]interface{}) {
			status = append(status, fmt.Sprintf("%s", str))
		}
	}

	if pm, ok := params[reporterpb.ParamsFieldPaymentMethod]; ok && pm != nil {
		for _, str := range pm.([]interface{}) {
			paymentMethods = append(paymentMethods, fmt.Sprintf("%s", str))
		}
	}

	if df, ok := params[reporterpb.ParamsFieldDateFrom]; ok {
		dateFrom = int64(df.(float64))
	}

	if dt, ok := params[reporterpb.ParamsFieldDateTo]; ok {
		dateTo = int64(dt.(float64))
	}

	ordersRequest := &billingpb.ListOrdersRequest{
		Merchant:      []string{h.report.MerchantId},
		Status:        status,
		PaymentMethod: paymentMethods,
		PmDateFrom:    dateFrom,
		PmDateTo:      dateTo,
	}
	orders, err := h.billing.FindAllOrdersPublic(ctx, ordersRequest)

	if err != nil || orders.Status != billingpb.ResponseStatusOk {
		if err == nil {
			err = errors.New(orders.Message.Message)
		}

		zap.L().Error(
			"Unable to get orders",
			zap.Error(err),
			zap.Any("request", ordersRequest),
		)

		return nil, err
	}

	for _, transaction := range orders.Item.Items {
		product := "Checkout"

		if len(transaction.Items) > 0 {
			if len(transaction.Items) == 1 {
				product = transaction.Items[0].Name
			} else {
				product = "Product"
			}
		}

		createdAt, err := ptypes.Timestamp(transaction.CreatedAt)

		if err != nil {
			zap.L().Error(
				"Unable to cast timestamp to time",
				zap.Error(err),
				zap.String("created_at", transaction.CreatedAt.String()),
			)
			return nil, err
		}

		logs = append(logs, map[string]interface{}{
			"project_name":   transaction.Project.Name["en"],
			"product_name":   product,
			"datetime":       createdAt.Format("2006-01-02T15:04:05"),
			"country":        transaction.CountryCode,
			"payment_method": transaction.PaymentMethod.Name,
			"transaction_id": transaction.Transaction,
			"net_amount":     math.Round(transaction.TotalPaymentAmount*100) / 100,
			"status":         transaction.Status,
			"currency":       transaction.Currency,
		})
	}

	reports := map[string]interface{}{
		"transactions": logs,
	}

	return reports, nil
}

func (h *Transactions) PostProcess(ctx context.Context, id, fileName string, retentionTime int64, content []byte) error {
	return nil
}
