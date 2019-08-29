package errors

import (
	"github.com/paysuper/paysuper-reporter/pkg/proto"
)

var (
	ErrorReportTypeNotFound           = newErrorMsg("rf000001", "unsupported report type.")
	ErrorFileTypeNotFound             = newErrorMsg("rf000002", "invalid file type.")
	ErrorUnableToCreate               = newErrorMsg("rf000003", "unable to create report file.")
	ErrorTemplateNotFound             = newErrorMsg("rf000004", "unable to update report file.")
	ErrorNotFound                     = newErrorMsg("rf000005", "report file not found.")
	ErrorCentrifugoNotificationFailed = newErrorMsg("rf000006", "unable to send report file to centrifugo.")
	ErrorMessageBrokerFailed          = newErrorMsg("rf000007", "unable to publish report file message to the message broker.")
	ErrorDocumentGeneratorRender      = newErrorMsg("rf000008", "document generator api return not success http status.")
	ErrorHandlerNotFound              = newErrorMsg("rf000009", "handler not found.")
	ErrorHandlerValidation            = newErrorMsg("rf000010", "handler validation error.")
	ErrorConvertBson                  = newErrorMsg("rf000011", "unable to convert report to bson.")
	ErrorParamIdNotFound              = newErrorMsg("rf000012", "report ID is not found.")
	ErrorDatabaseQueryFailed          = newErrorMsg("rf000013", "Query to database collection failed")
)

func newErrorMsg(code, msg string, details ...string) *proto.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return &proto.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}
