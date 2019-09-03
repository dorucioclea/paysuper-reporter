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

// NatsConfig defines the parameters for connecting to the NATS streaming server.
type NatsConfig struct {
	ServerUrls string `envconfig:"NATS_SERVER_URLS" default:"127.0.0.1:4222"`
	ClusterId  string `envconfig:"NATS_CLUSTER_ID" default:"test-cluster"`
	ClientId   string `envconfig:"NATS_CLIENT_ID" default:"billing-server-publisher"`
	Async      bool   `envconfig:"NATS_ASYNC" default:"false"`
	User       string `envconfig:"NATS_USER" default:""`
	Password   string `envconfig:"NATS_PASSWORD" default:""`
}

// AWS defines the parameters for connecting to the NATS streaming server.
type S3Config struct {
	AccessKeyId string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	SecretKey   string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	Region      string `envconfig:"AWS_REGION" required:"true"`
	BucketName  string `envconfig:"AWS_BUCKET" required:"true"`
}

// Centrifugo defines the parameters for connecting to the Centrifugo server.
type CentrifugoConfig struct {
	ApiSecret       string `envconfig:"CENTRIFUGO_API_SECRET" required:"true"`
	Secret          string `envconfig:"CENTRIFUGO_SECRET" required:"true"`
	URL             string `envconfig:"CENTRIFUGO_URL" required:"false" default:"http://127.0.0.1:8000"`
	MerchantChannel string `envconfig:"CENTRIFUGO_MERCHANT_CHANNEL" default:"paysuper:merchant#%s"`
}

// DocumentGeneratorConfig defines the parameters for connecting to the document generator service.
type DocumentGeneratorConfig struct {
	ApiUrl  string `envconfig:"DOCGEN_API_URL" default:"http://127.0.0.1:5488"`
	Timeout int    `envconfig:"DOCGEN_API_TIMEOUT" default:"60000"`
}

type Config struct {
	Db               MongoConfig
	Nats             NatsConfig
	S3               S3Config
	DG               DocumentGeneratorConfig
	CentrifugoConfig CentrifugoConfig

	MetricsPort           string `envconfig:"METRICS_PORT" required:"false" default:"8086"`
	MicroRegistry         string `envconfig:"MICRO_REGISTRY" required:"false"`
	MicroSelector         string `envconfig:"MICRO_SELECTOR" required:"false" default:"static"`
	DocumentRetentionTime string `envconfig:"DOCUMENT_RETENTION_TIME" default:"604800"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	return cfg, err
}
