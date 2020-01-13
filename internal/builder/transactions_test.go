package builder

import (
	"encoding/json"
	errs "errors"
	"github.com/globalsign/mgo/bson"
	billingProto "github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/internal/mocks"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type TransactionsBuilderTestSuite struct {
	suite.Suite
	service BuildInterface
}

func Test_TransactionsBuilder(t *testing.T) {
	suite.Run(t, new(TransactionsBuilderTestSuite))
}

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Validate_Error_MerchantIdNotFound() {
	params, _ := json.Marshal(map[string]interface{}{})
	h := newTransactionsHandler(&Handler{
		report: &reporterpb.ReportFile{Params: params},
	})

	assert.Errorf(suite.T(), h.Validate(), errors.ErrorParamMerchantIdNotFound.Message)
}

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Validate_Ok() {
	h := newTransactionsHandler(&Handler{
		report: &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex()},
	})

	assert.NoError(suite.T(), h.Validate())
}

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Build_Error_FindByMerchant() {
	rep := mocks.TransactionsRepositoryInterface{}
	rep.
		On("FindByMerchant", mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything).
		Return(nil, errs.New("not found"))

	params, _ := json.Marshal(map[string]interface{}{})
	h := newTransactionsHandler(&Handler{
		transactionsRepository: &rep,
		report:                 &reporterpb.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Build_Ok() {
	var status []string
	var paymentMethods []string

	merchantId := bson.NewObjectId().Hex()

	rep := mocks.TransactionsRepositoryInterface{}
	rep.
		On("FindByMerchant", merchantId, status, paymentMethods, int64(0), int64(0)).
		Return([]*billingProto.MgoOrderViewPublic{
			{
				Project:            &billingProto.MgoOrderProject{Name: []*billingProto.MgoMultiLang{{Value: "name"}}},
				CreatedAt:          time.Now(),
				CountryCode:        "RU",
				PaymentMethod:      &billingProto.MgoOrderPaymentMethod{Name: "payment"},
				Transaction:        "123123",
				TotalPaymentAmount: float64(123),
				Status:             "status",
				Currency:           "RUB",
			},
		}, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newTransactionsHandler(&Handler{
		transactionsRepository: &rep,
		report:                 &reporterpb.ReportFile{MerchantId: merchantId, Params: params},
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Build_Ok_CustomParams() {
	rep := mocks.TransactionsRepositoryInterface{}
	rep.
		On("FindByMerchant", mock2.Anything, []string{"processed"}, []string{"card", "qiwi"}, int64(1571225221), int64(1573817221)).
		Return([]*billingProto.MgoOrderViewPublic{
			{
				Project:            &billingProto.MgoOrderProject{Name: []*billingProto.MgoMultiLang{{Value: "name"}}},
				CreatedAt:          time.Now(),
				CountryCode:        "RU",
				PaymentMethod:      &billingProto.MgoOrderPaymentMethod{Name: "payment"},
				Transaction:        "123123",
				TotalPaymentAmount: float64(123),
				Status:             "status",
				Currency:           "RUB",
			},
		}, nil)

	params, _ := json.Marshal(map[string]interface{}{
		pkg.ParamsFieldStatus:        []interface{}{"processed"},
		pkg.ParamsFieldPaymentMethod: []interface{}{"card", "qiwi"},
		pkg.ParamsFieldDateFrom:      1571225221,
		pkg.ParamsFieldDateTo:        1573817221,
	})
	h := newTransactionsHandler(&Handler{
		transactionsRepository: &rep,
		report:                 &reporterpb.ReportFile{Params: params},
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}
