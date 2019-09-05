package pkg

const (
	ServiceName    = "p1payreporter"
	ServiceVersion = "latest"

	LoggerName = "PAYSUPER_BILLING_REPORTER"

	SubjectRequestReportFileCreate = "report_file_create"
	FileMask                       = "report_%s.%s"

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

	OutputXlsxExtension = "xlsx"
	OutputCsvExtension  = "csv"
	OutputPdfExtension  = "pdf"

	OutputXlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	OutputCsvContentType  = "text/csv"
	OutputPdfContentType  = "application/pdf"

	ParamsFieldId = "id"
)
