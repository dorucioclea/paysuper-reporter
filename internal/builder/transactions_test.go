package builder

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	billingMocks "github.com/paysuper/paysuper-proto/go/billingpb/mocks"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
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

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Build_Ok() {
	billing := &billingMocks.BillingService{}

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status: billingpb.ResponseStatusOk,
		Item: &billingpb.ListOrdersPublicResponseItem{
			Items: suite.getOrdersTemplate(),
		},
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.NoError(suite.T(), err)
}

func (suite *TransactionsBuilderTestSuite) TestTransactionsBuilder_Build_Error_FindAllOrdersPublic() {
	billing := &billingMocks.BillingService{}

	ordersResponse := &billingpb.ListOrdersPublicResponse{
		Status:  billingpb.ResponseStatusNotFound,
		Message: &billingpb.ResponseErrorMessage{Message: "error"},
		Item:    nil,
	}
	billing.On("FindAllOrdersPublic", mock2.Anything, mock2.Anything).Return(ordersResponse, nil)

	params, _ := json.Marshal(map[string]interface{}{})
	h := newTransactionsHandler(&Handler{
		report:  &reporterpb.ReportFile{MerchantId: bson.NewObjectId().Hex(), Params: params},
		billing: billing,
	})

	_, err := h.Build()
	assert.Error(suite.T(), err)
}

func (suite *TransactionsBuilderTestSuite) getOrdersTemplate() []*billingpb.OrderViewPublic {
	datetime, _ := ptypes.TimestampProto(time.Now())

	return []*billingpb.OrderViewPublic{
		{
			Project:            &billingpb.ProjectOrder{Name: map[string]string{"en": "name"}},
			CreatedAt:          datetime,
			CountryCode:        "RU",
			PaymentMethod:      &billingpb.PaymentMethodOrder{Name: "payment"},
			Transaction:        "123123",
			TotalPaymentAmount: float64(123),
			Status:             "status",
			Currency:           "RUB",
		},
	}
}
