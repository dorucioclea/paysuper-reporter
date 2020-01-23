package pkg

const (
	ServiceName    = "p1payreporter"
	ServiceVersion = "latest"

	LoggerName = "PAYSUPER_REPORTER"

	FileMask          = "report_%s_%s.%s"
	FileMaskAgreement = "License Agreement_%s_#%s.%s"

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
	ReportTypeAgreement           = "agreement"

	OutputExtensionXlsx = "xlsx"
	OutputExtensionCsv  = "csv"
	OutputExtensionPdf  = "pdf"

	OutputContentTypeXlsx = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	OutputContentTypeCsv  = "text/csv"
	OutputContentTypePdf  = "application/pdf"

	RecipeXlsx = "html-to-xlsx"
	RecipeCsv  = "text"
	RecipePdf  = "chrome-pdf"

	ParamsFieldId            = "id"
	ParamsFieldCountry       = "country"
	ParamsFieldStatus        = "status"
	ParamsFieldPaymentMethod = "payment_method"
	ParamsFieldDateFrom      = "date_from"
	ParamsFieldDateTo        = "date_to"

	RequestParameterAgreementNumber                             = "number"
	RequestParameterAgreementLegalName                          = "legal_name"
	RequestParameterAgreementAddress                            = "address"
	RequestParameterAgreementRegistrationNumber                 = "registration_number"
	RequestParameterAgreementPayoutCost                         = "payout_cost"
	RequestParameterAgreementMinimalPayoutLimit                 = "minimal_payout_limit"
	RequestParameterAgreementPayoutCurrency                     = "payout_currency"
	RequestParameterAgreementPSRate                             = "ps_rate"
	RequestParameterAgreementHomeRegion                         = "home_region"
	RequestParameterAgreementMerchantAuthorizedName             = "merchant_authorized_name"
	RequestParameterAgreementMerchantAuthorizedPosition         = "merchant_authorized_position"
	RequestParameterAgreementOperatingCompanyLegalName          = "oc_name"
	RequestParameterAgreementOperatingCompanyAddress            = "oc_address"
	RequestParameterAgreementOperatingCompanyRegistrationNumber = "oc_registration_number"
	RequestParameterAgreementOperatingCompanyAuthorizedName     = "oc_authorized_name"
	RequestParameterAgreementOperatingCompanyAuthorizedPosition = "oc_authorized_position"

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
