Document reporter
=====

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-brightgreen.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Build Status](https://travis-ci.org/paysuper/paysuper-reporter.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-reporter) 
[![codecov](https://codecov.io/gh/paysuper/paysuper-reporter/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-reporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-reporter)](https://goreportcard.com/report/github.com/paysuper/paysuper-reporter)

Consumer NAT Streaming server to generate reporting documents (royalties, vats, transactions and others).

## Environment variables:

| Name                            | Required | Default                                        | Description                                                                                                                             |
|:--------------------------------|:--------:|:-----------------------------------------------|:------------------------------------------------------------------------|
| METRICS_PORT                    | -        | 8086                                           | Http server port for health and metrics request                         |
| MICRO_SELECTOR                  | -        | static                                         | Type of selector for Micro service                                      |
| MONGO_DSN                       | true     | -                                              | MongoBD DSN connection string                                           |
| MONGO_DIAL_TIMEOUT              | -        | 10                                             | MongoBD dial timeout in seconds                                         |
| MONGO_MODE                      | -        | 4                                              | Consistency mode for the MongoDB session                                |
| NATS_SERVER_URLS                | -        | 127.0.0.1:4222                                 | The nats server URLs (separated by comma)                               |
| NATS_ASYNC                      | -        | false                                          | Publish asynchronously                                                  |
| NATS_USER                       | -        |                                                | User sets the username to be used when connecting to the server         |
| NATS_PASSWORD                   | -        |                                                | Password sets the password to be used when connecting to a server       |
| NATS_CLUSTER_ID                 | -        | test-cluster                                   | The NATS Streaming cluster ID                                           |
| NATS_CLIENT_ID                  | -        | billing-server-publisher                       | The NATS Streaming client ID to connect with                            |
| AWS_ACCESS_KEY_ID               | true     |                                                |                                                                         |
| AWS_SECRET_ACCESS_KEY           | true     |                                                |                                                                         |
| AWS_TOKEN                       | true     |                                                |                                                                         |
| AWS_BUCKET                      | true     |                                                |                                                                         |
| AWS_REGION                      | true     |                                                |                                                                         |
| CENTRIFUGO_SECRET               | true     | -                                              | Centrifugo secret key                                                   |
| CENTRIFUGO_API_SECRET           | true     | -                                              | Centrifugo API secret key                                               |
| CENTRIFUGO_URL                  | -        | http://127.0.0.1:8000                          | Centrifugo API gateway                                                  |
| CENTRIFUGO_MERCHANT_CHANNEL     | -        | paysuper:merchant#%s                           | Centrifugo channel name to send notifications to merchant               |
| DOCGEN_API_URL                  | -        | http://127.0.0.1:5488                          | URL of document generation service                                      |
| DOCGEN_API_TIMEOUT              | -        | 60000                                          | Timeout for waiting for a response from the document generation service |
| DOCUMENT_RETENTION_TIME         | -        | 604800                                         | Time to live the document in the S3 and DB storage                      |

## Usage Example:

```go
package main

import (
    reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
    "github.com/streadway/amqp"
    "gopkg.in/ProtocolONE/rabbitmq.v1/pkg"
    "log"
)

func main()  {
    broker, err := rabbitmq.NewBroker("amqp://127.0.0.1:5672")
    
    if err != nil {
        log.Fatalf("Creating RabbitMQ publisher failed with error: %s\n", err)
    }
    
    payload := &reporterPkg.Payload{
        TemplateAlias: "template_name",
        TemplateModel: map[string]string{"param1": "value1"},
        To:            "emai@test.com",
    }
    
    err = broker.Publish(reporterPkg.PostmarkSenderTopicName, payload, amqp.Table{})
    
    if err != nil {
        log.Fatalf("Publication message to queue failed with error: %s\n", err)
    }
}
```

## Contributing
We feel that a welcoming community is important and we ask that you follow PaySuper's [Open Source Code of Conduct](https://github.com/paysuper/code-of-conduct/blob/master/README.md) in all interactions with the community.

PaySuper welcomes contributions from anyone and everyone. Please refer to each project's style and contribution guidelines for submitting patches and additions. In general, we follow the "fork-and-pull" Git workflow.

The master branch of this repository contains the latest stable release of this component.