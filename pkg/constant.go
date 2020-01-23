package pkg

const (
	LoggerName = "PAYSUPER_REPORTER"

	MIMEApplicationJSON = "application/json"

	ResponseStatusOk          = int32(200)
	ResponseStatusBadData     = int32(400)
	ResponseStatusSystemError = int32(500)

	OutputContentTypeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	OutputContentTypeCsv  = "text/csv"
	OutputContentTypePdf  = "application/pdf"

	RecipeXlsx = "html-to-xlsx"
	RecipeCsv  = "text"
	RecipePdf  = "chrome-pdf"

	BrokerMessageRetryMaxCount = 10

	BrokerGenerateReportTopicName = "reporter-generate"
	BrokerPostProcessTopicName    = "reporter-post-process"
)
