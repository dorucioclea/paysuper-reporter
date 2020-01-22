module github.com/paysuper/paysuper-reporter

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/ProtocolONE/nats v0.0.0-20190909153110-738ec68e5d7c
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/golang-migrate/migrate/v4 v4.6.2
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/now v1.0.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/micro/go-micro v1.8.0
	github.com/micro/go-plugins v1.2.0
	github.com/mongodb/mongo-go-driver v0.3.0
	github.com/nats-io/stan.go v0.5.0
	github.com/paysuper/paysuper-aws-manager v0.0.1
	github.com/paysuper/paysuper-billing-server v0.0.0-20191114134535-c158b5075a9a
	github.com/paysuper/paysuper-database-mongo v0.1.3
	github.com/paysuper/paysuper-recurring-repository v1.0.126
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/stretchr/testify v1.4.0
	go.mongodb.org/mongo-driver v1.1.3
	go.uber.org/zap v1.10.0
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
