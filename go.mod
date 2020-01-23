module github.com/paysuper/paysuper-reporter

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/aws/aws-sdk-go v1.23.8
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/golang-migrate/migrate/v4 v4.6.2
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/now v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.2.0
	github.com/paysuper/paysuper-aws-manager v0.0.1
	github.com/paysuper/paysuper-billing-server v1.1.0
	github.com/paysuper/paysuper-proto/go/billingpb v0.0.0-20200122212511-cb73c60b18e4
	github.com/paysuper/paysuper-proto/go/reporterpb v0.0.0-20200122212511-cb73c60b18e4
	github.com/paysuper/paysuper-tools v0.0.0-20200117101901-522574ce4d1c
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.4.0
	go.mongodb.org/mongo-driver v1.2.1
	go.uber.org/zap v1.13.0
	gopkg.in/ProtocolONE/rabbitmq.v1 v1.0.0-20191130200733-22b27ffa73aa
	gopkg.in/paysuper/paysuper-database-mongo.v2 v2.0.0-20200116095540-a477bfd0ce4c
)

replace (
	github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	github.com/paysuper/paysuper-tax-service => github.com/paysuper/paysuper-tax-service v1.0.0
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190927073244-c990c680b611
)

go 1.13
