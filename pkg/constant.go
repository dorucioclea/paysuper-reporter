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

	ReportTypeTax             = "tax"
	ReportTypeTaxTemplate     = "tax_report"
	ReportTypeVat             = "vat"
	ReportTypeVatTemplate     = "vat_report"
	ReportTypeRoyalty         = "royalty"
	ReportTypeRoyaltyTemplate = "royalty_report"

	OutputXlsxExtension = "xlsx"
	OutputCsvExtension  = "csv"
	OutputPdfExtension  = "pdf"

	OutputXlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	OutputCsvContentType  = "text/csv"
	OutputPdfContentType  = "application/pdf"
)
