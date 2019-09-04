package pkg

const (
	ServiceName    = "p1payreporter"
	ServiceVersion = "latest"

	LoggerName = "PAYSUPER_BILLING_REPORTER"

	SubjectRequestReportFileCreate = "report_file_create"

	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"

	MIMEApplicationJSON = "application/json"

	ResponseStatusOk          = int32(200)
	ResponseStatusBadData     = int32(400)
	ResponseStatusNotFound    = int32(404)
	ResponseStatusSystemError = int32(500)

	ReportTypeTransactions                = "transactions"
	ReportTypeTransactionsTemplate        = "transactions_report"
	ReportTypeVat                         = "vat"
	ReportTypeVatTemplate                 = "vat_report"
	ReportTypeVatTransactions             = "vat_transactions"
	ReportTypeVatTransactionsTemplate     = "vat_transactions_report"
	ReportTypeRoyalty                     = "royalty"
	ReportTypeRoyaltyTemplate             = "royalty_report"
	ReportTypeRoyaltyTransactions         = "royalty_transactions"
	ReportTypeRoyaltyTransactionsTemplate = "royalty_transactions_report"

	OutputXlsxExtension = "xlsx"
	OutputCsvExtension  = "csv"
	OutputPdfExtension  = "pdf"

	OutputXlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	OutputCsvContentType  = "text/csv"
	OutputPdfContentType  = "application/pdf"

	ParamsFieldId = "id"
)
