package errors

import (
	"github.com/paysuper/paysuper-reporter/pkg/proto"
)

var (
	ErrorTemplateNotFound             = newErrorMsg("rf000001", "could not find a template for the report.")
	ErrorType                         = newErrorMsg("rf000002", "invalid file type.")
	ErrorUnableToCreate               = newErrorMsg("rf000003", "unable to create report file.")
	ErrorUnableToUpdate               = newErrorMsg("rf000004", "unable to update report file.")
	ErrorNotFound                     = newErrorMsg("rf000005", "report file not found.")
	ErrorCentrifugoNotificationFailed = newErrorMsg("rf000006", "unable to send report file to centrifugo.")
	ErrorMessageBrokerFailed          = newErrorMsg("rf000007", "unable to publish report file message to the message broker.")
	ErrorMarshalMatch                 = newErrorMsg("rf000008", "unable to marshal match data.")
	ErrorMarshalGroup                 = newErrorMsg("rf000009", "unable to marshal group data.")
	ErrorDocumentGeneratorRender      = newErrorMsg("rf000010", "document generator api return not success http status.")
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
