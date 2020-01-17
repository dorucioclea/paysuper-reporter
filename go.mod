module github.com/paysuper/paysuper-reporter

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/InVisionApp/go-logger v1.0.1 // indirect
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/golang-migrate/migrate/v4 v4.6.2
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/now v1.0.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.2.0
	github.com/paysuper/paysuper-aws-manager v0.0.1
	github.com/paysuper/paysuper-database-mongo v0.1.3
	github.com/paysuper/paysuper-proto/go/billingpb v0.0.0-20200116145615-427433ee02be
	github.com/paysuper/paysuper-proto/go/recurringpb v0.0.0-20200114235009-da02b724903d // indirect
	github.com/paysuper/paysuper-proto/go/reporterpb v0.0.0-20200117172130-df1a443c1fe8
	github.com/paysuper/paysuper-tools v0.0.0-20200115135413-15b9d03f5ec4
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.4.0
	github.com/vektra/mockery v0.0.0-20181123154057-e78b021dcbb5 // indirect
	go.uber.org/zap v1.13.0
	gopkg.in/ProtocolONE/rabbitmq.v1 v1.0.0-20191130200733-22b27ffa73aa
)

replace (
	github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190927073244-c990c680b611
)

go 1.13
