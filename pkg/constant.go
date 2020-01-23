package pkg

const (
	LoggerName = "PAYSUPER_REPORTER"

	FileMask          = "report_%s_%s.%s"
	FileMaskAgreement = "License Agreement_%s_#%s.%s"

	HeaderContentType = "Content-Type"

	MIMEApplicationJSON = "application/json"

	ResponseStatusOk          = int32(200)
	ResponseStatusBadData     = int32(400)
	ResponseStatusNotFound    = int32(404)
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

	ErrorDatabaseQueryFailed        = "Query to database collection failed"
	ErrorDatabaseInvalidObjectId    = "String is not a valid ObjectID"
	ErrorQueryCursorExecutionFailed = "Execute result from query cursor failed"
	ErrorDatabaseFieldCollection    = "collection"
	ErrorDatabaseFieldQuery         = "query"
	ErrorDatabaseFieldObjectId      = "object_id"
)
