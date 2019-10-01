module github.com/paysuper/paysuper-reporter

require (
	github.com/InVisionApp/go-health v2.1.0+incompatible
	github.com/ProtocolONE/nats v0.0.0-20190909153110-738ec68e5d7c
	github.com/centrifugal/gocent v2.0.2+incompatible
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/golang-migrate/migrate/v4 v4.3.1
	github.com/golang/protobuf v1.3.2
	github.com/jinzhu/now v1.0.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/micro/go-micro v1.8.0
	github.com/micro/go-plugins v1.2.0
	github.com/nats-io/stan.go v0.5.0
	github.com/paysuper/paysuper-aws-manager v0.0.1
	github.com/paysuper/paysuper-billing-server v0.0.0-20190927123432-893bab0748f1
	github.com/paysuper/paysuper-database-mongo v0.1.1
	github.com/paysuper/paysuper-recurring-repository v1.0.123
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
)

replace (
	github.com/gogo/protobuf v0.0.0-20190410021324-65acae22fc9 => github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
	golang.org/x/sys => golang.org/x/sys v0.0.0-20190927073244-c990c680b611
)

go 1.13
