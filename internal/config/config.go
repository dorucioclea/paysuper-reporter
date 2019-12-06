package config

import (
	"github.com/globalsign/mgo"
	"github.com/kelseyhightower/envconfig"
)

// MongoConfig defines the parameters for connecting to the MongoDB.
type MongoConfig struct {
	Dsn         string   `envconfig:"MONGO_DSN" required:"true"`
	DialTimeout string   `envconfig:"MONGO_DIAL_TIMEOUT" required:"false" default:"10"`
	MongoMode   mgo.Mode `envconfig:"MONGO_MODE" required:"false" default:"4"`
}

// AWS defines the parameters for connecting to the NATS streaming server.
type S3Config struct {
	AccessKeyId string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretKey   string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	Region      string `envconfig:"AWS_REGION" required:"true"`
	BucketName  string `envconfig:"AWS_BUCKET" required:"true"`

	AwsAccessKeyIdAgreement     string `envconfig:"AWS_ACCESS_KEY_ID_AGREEMENT" required:"true"`
	AwsSecretAccessKeyAgreement string `envconfig:"AWS_SECRET_ACCESS_KEY_AGREEMENT" required:"true"`
	AwsRegionAgreement          string `envconfig:"AWS_REGION_AGREEMENT" default:"eu-west-1"`
	AwsBucketAgreement          string `envconfig:"AWS_BUCKET_AGREEMENT" required:"true"`
}

// Centrifugo defines the parameters for connecting to the Centrifugo server.
type CentrifugoConfig struct {
	ApiSecret   string `envconfig:"CENTRIFUGO_API_SECRET" required:"true"`
	URL         string `envconfig:"CENTRIFUGO_URL" required:"false" default:"http://127.0.0.1:8000"`
	UserChannel string `envconfig:"CENTRIFUGO_USER_CHANNEL" default:"paysuper:user#%s"`
}

// DocumentGeneratorConfig defines the parameters for connecting to the document generator service.
type DocumentGeneratorConfig struct {
	ApiUrl                      string `envconfig:"DOCGEN_API_URL" default:"http://127.0.0.1:5488"`
	Timeout                     int    `envconfig:"DOCGEN_API_TIMEOUT" default:"60000"`
	Username                    string `envconfig:"DOCGEN_USERNAME" default:""`
	Password                    string `envconfig:"DOCGEN_PASSWORD" default:""`
	RoyaltyTemplate             string `envconfig:"DOCGEN_ROYALTY_TEMPLATE" required:"true"`
	RoyaltyTransactionsTemplate string `envconfig:"DOCGEN_ROYALTY_TRANSACTIONS_TEMPLATE" required:"true"`
	VatTemplate                 string `envconfig:"DOCGEN_VAT_TEMPLATE" required:"true"`
	VatTransactionsTemplate     string `envconfig:"DOCGEN_VAT_TRANSACTIONS_TEMPLATE" required:"true"`
	TransactionsTemplate        string `envconfig:"DOCGEN_TRANSACTIONS_TEMPLATE" required:"true"`
	PayoutTemplate              string `envconfig:"DOCGEN_PAYOUT_TEMPLATE" required:"true"`
	AgreementTemplate           string `envconfig:"DOCGEN_AGREEMENT_TEMPLATE" required:"true"`
}

type Config struct {
	Db               MongoConfig
	S3               S3Config
	DG               DocumentGeneratorConfig
	CentrifugoConfig CentrifugoConfig

	MetricsPort           string `envconfig:"METRICS_PORT" required:"false" default:"8086"`
	MicroSelector         string `envconfig:"MICRO_SELECTOR" required:"false" default:"static"`
	DocumentRetentionTime int64  `envconfig:"DOCUMENT_RETENTION_TIME" default:"604800"`
	BrokerAddress         string `envconfig:"BROKER_ADDRESS" default:"amqp://127.0.0.1:5672"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	return cfg, err
}
