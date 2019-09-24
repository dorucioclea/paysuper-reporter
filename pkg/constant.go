package pkg

const (
	ServiceName    = "p1payreporter"
	ServiceVersion = "latest"

	LoggerName = "PAYSUPER_BILLING_REPORTER"

	SubjectRequestReportFileCreate = "report_file_create"
	FileMask                       = "report_%s_%s.%s"

	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"

	MIMEApplicationJSON = "application/json"

	ResponseStatusOk          = int32(200)
	ResponseStatusBadData     = int32(400)
	ResponseStatusNotFound    = int32(404)
	ResponseStatusSystemError = int32(500)

	ReportTypeTransactions        = "transactions"
	ReportTypeVat                 = "vat"
	ReportTypeVatTransactions     = "vat_transactions"
	ReportTypeRoyalty             = "royalty"
	ReportTypeRoyaltyTransactions = "royalty_transactions"
	ReportTypePayout              = "payout"

	OutputExtensionXlsx = "html-to-xlsx"
	OutputExtensionCsv  = "csv"
	OutputExtensionPdf  = "pdf"

	OutputContentTypeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	OutputContentTypeCsv  = "text/csv"
	OutputContentTypePdf  = "application/pdf"

	RecipeXlsx = "xlsx"
	RecipeCsv  = "text"
	RecipePdf  = "chrome-pdf"

	ParamsFieldId = "id"
)
