package config

import "github.com/kelseyhightower/envconfig"

// MongoConfig defines the parameters for connecting to the MongoDB.
type MongoConfig struct {
	Dsn         string `envconfig:"MONGO_DSN" required:"true"`
	DialTimeout string `envconfig:"MONGO_DIAL_TIMEOUT" required:"false" default:"10"`
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

// NatsConfig defines the parameters for connecting to the NATS streaming server.
type S3Config struct {
	AccessKeyId string `envconfig:"S3_ACCESS_KEY" required:"true"`
	SecretKey   string `envconfig:"S3_SECRET_KEY" required:"true"`
	Endpoint    string `envconfig:"S3_ENDPOINT" required:"true"`
	BucketName  string `envconfig:"S3_BUCKET_NAME" required:"true"`
	Region      string `envconfig:"S3_REGION" default:"us-west-2"`
	Secure      bool   `envconfig:"S3_SECURE" default:"false"`
}

// DocumentGenerator defines the parameters for connecting to the document generator service.
type DocumentGenerator struct {
	ApiUrl  string `envconfig:"DOC_API_URL" default:"http://127.0.0.1:5488"`
	Timeout int    `envconfig:"DOC_API_URL" default:"60000"`
}

type Config struct {
	Db   MongoConfig
	Nats NatsConfig
	S3   S3Config
	DG   DocumentGenerator

	MicroRegistry string `envconfig:"MICRO_REGISTRY" required:"false"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process("", cfg)

	return cfg, err
}
