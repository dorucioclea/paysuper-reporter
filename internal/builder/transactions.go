package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math"
	"reflect"
)

type Transactions DefaultHandler

func newTransactionsHandler(h *Handler) BuildInterface {
	return &Transactions{Handler: h}
}

func (h *Transactions) Validate() error {
	params, _ := h.GetParams()
	_, err := primitive.ObjectIDFromHex(h.report.MerchantId)

	if err != nil {
		return errors.New(errs.ErrorParamMerchantIdNotFound.Message)
	}

	if st, ok := params[pkg.ParamsFieldStatus]; ok && st != nil {
		if reflect.TypeOf(st).Kind() != reflect.Slice {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}
	}

	if st, ok := params[pkg.ParamsFieldPaymentMethod]; ok && st != nil {
		if reflect.TypeOf(st).Kind() != reflect.Slice {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}

		for _, str := range st.([]interface{}) {
			_, err = primitive.ObjectIDFromHex(fmt.Sprintf("%s", str))

			if err != nil {
				return errors.New(errs.ErrorHandlerValidation.Message)
			}
		}
	}

	if st, ok := params[pkg.ParamsFieldDateFrom]; ok {
		if reflect.TypeOf(st).Kind() != reflect.Float64 {
			return errors.New(errs.ErrorHandlerValidation.Message)
		}
	}

	if st, ok := params[pkg.ParamsFieldDateTo]; ok {
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

	params, _ := h.GetParams()

	if st, ok := params[pkg.ParamsFieldStatus]; ok && st != nil {
		for _, str := range st.([]interface{}) {
			status = append(status, fmt.Sprintf("%s", str))
		}
	}

	if pm, ok := params[pkg.ParamsFieldPaymentMethod]; ok && pm != nil {
		for _, str := range pm.([]interface{}) {
			paymentMethods = append(paymentMethods, fmt.Sprintf("%s", str))
		}
	}

	if df, ok := params[pkg.ParamsFieldDateFrom]; ok {
		dateFrom = int64(df.(float64))
	}

	if dt, ok := params[pkg.ParamsFieldDateTo]; ok {
		dateTo = int64(dt.(float64))
	}

	transactions, err := h.transactionsRepository.FindByMerchant(h.report.MerchantId, status, paymentMethods, dateFrom, dateTo)

	if err != nil {
		return nil, err
	}

	for _, transaction := range transactions {
		product := "Checkout"

		if len(transaction.Items) > 0 {
			if len(transaction.Items) == 1 {
				product = transaction.Items[0].Name
			} else {
				product = "Product"
			}
		}

		logs = append(logs, map[string]interface{}{
			"project_name":   transaction.Project.Name[0].Value,
			"product_name":   product,
			"datetime":       transaction.CreatedAt.Format("2006-01-02T15:04:05"),
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

func (h *Transactions) PostProcess(_ context.Context, _, _ string, _ int64, _ []byte) error {
	return nil
}
