package errors

import (
	"github.com/paysuper/paysuper-proto/go/reporterpb"
)

var (
	ErrorReportTypeNotFound           = newErrorMsg("rf000001", "unsupported report type.")
	ErrorUnableToCreate               = newErrorMsg("rf000003", "unable to create report file.")
	ErrorTemplateNotFound             = newErrorMsg("rf000004", "template not found.")
	ErrorFileType                     = newErrorMsg("rf000005", "invalid file type.")
	ErrorCentrifugoNotificationFailed = newErrorMsg("rf000006", "unable to send report file to centrifugo.")
	ErrorMessageBrokerFailed          = newErrorMsg("rf000007", "unable to publish report file message to the message broker.")
	ErrorDocumentGeneratorRender      = newErrorMsg("rf000008", "document generator api return not success http status.")
	ErrorHandlerNotFound              = newErrorMsg("rf000009", "handler not found.")
	ErrorHandlerValidation            = newErrorMsg("rf000010", "handler validation error.")
	ErrorParamCountryNotFound         = newErrorMsg("rf000011", "unable to find the param <country>.")
	ErrorParamIdNotFound              = newErrorMsg("rf000012", "unable to find the param <id>.")
	ErrorParamMerchantIdNotFound      = newErrorMsg("rf000012", "unable to find the param <merchant_id>.")
	ErrorDatabaseQueryFailed          = newErrorMsg("rf000013", "query to database collection failed")
	ErrorMongoDbOidIncorrect          = newErrorMsg("rf000014", "mongodb object id incorrect")
)

func newErrorMsg(code, msg string, details ...string) *reporterpb.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return &reporterpb.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}
