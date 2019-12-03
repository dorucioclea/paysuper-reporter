Document reporter
=====

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-brightgreen.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Build Status](https://travis-ci.org/paysuper/paysuper-reporter.svg?branch=master)](https://travis-ci.org/paysuper/paysuper-reporter) 
[![codecov](https://codecov.io/gh/paysuper/paysuper-reporter/branch/master/graph/badge.svg)](https://codecov.io/gh/paysuper/paysuper-reporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/paysuper/paysuper-reporter)](https://goreportcard.com/report/github.com/paysuper/paysuper-reporter)

DocumentReporter is a GRPS service for creating printable reports (royalties, vats, transactions, etc.).

## Environment variables:

| Name                                 | Required | Default                                        | Description                                                                                                                             |
|:-------------------------------------|:--------:|:-----------------------------------------------|:------------------------------------------------------------------------|
| METRICS_PORT                         | -        | 8086                                           | Http server port for health and metrics request                         |
| MICRO_SELECTOR                       | -        | static                                         | Type of selector for Micro service                                      |
| MONGO_DSN                            | true     | -                                              | MongoBD DSN connection string                                           |
| MONGO_DIAL_TIMEOUT                   | -        | 10                                             | MongoBD dial timeout in seconds                                         |
| MONGO_MODE                           | -        | 4                                              | Consistency mode for the MongoDB session                                |
| BROKER_ADDRESS                       | -        | amqp://127.0.0.1:5672                          | RabbitMQ url address                                                    |
| AWS_ACCESS_KEY_ID                    | true     |                                                |                                                                         |
| AWS_SECRET_ACCESS_KEY                | true     |                                                |                                                                         |
| AWS_BUCKET                           | true     |                                                |                                                                         |
| AWS_REGION                           | true     |                                                |                                                                         |
| AWS_ACCESS_KEY_ID_AGREEMENT          | true     | -                                              | AWS access key identifier for agreements storage                        |
| AWS_SECRET_ACCESS_KEY_AGREEMENT      | true     | -                                              | AWS access secret key for agreements storage                            |
| AWS_BUCKET_AGREEMENT                 | true     | -                                              | AWS bucket name for agreements storage                                  |
| AWS_REGION_AGREEMENT                 | -        | eu-west-1                                      | AWS region for agreements storage                                       |
| CENTRIFUGO_API_SECRET                | true     | -                                              | Centrifugo API secret key                                               |
| CENTRIFUGO_URL                       | -        | http://127.0.0.1:8000                          | Centrifugo API gateway                                                  |
| CENTRIFUGO_USER_CHANNEL              | -        | paysuper:user#%s                               | Centrifugo channel name to send notifications to user                   |
| DOCGEN_API_URL                       | -        | http://127.0.0.1:5488                          | URL of document generation service                                      |
| DOCGEN_API_TIMEOUT                   | -        | 60000                                          | Timeout for waiting for a response from the document generation service |
| DOCGEN_USERNAME                      | -        |                                                | Username for authenticate                                               |
| DOCGEN_PASSWORD                      | -        |                                                | Password for authenticate                                               |
| DOCGEN_ROYALTY_TEMPLATE              | true     |                                                | ID of template in the JSReport for royalty report                       |
| DOCGEN_ROYALTY_TRANSACTIONS_TEMPLATE | true     |                                                | ID of template in the JSReport for royalty transactions report          |
| DOCGEN_VAT_TEMPLATE                  | true     |                                                | ID of template in the JSReport for vat report                           |
| DOCGEN_VAT_TRANSACTIONS_TEMPLATE     | true     |                                                | ID of template in the JSReport for vat transactions report              |
| DOCGEN_TRANSACTIONS_TEMPLATE         | true     |                                                | ID of template in the JSReport for find transactions report             |
| DOCGEN_PAYOUT_TEMPLATE               | true     |                                                | ID of template in the JSReport for payout report                        |
| DOCGEN_AGREEMENT_TEMPLATE            | true     |                                                | ID of template in the JSReport for merchant agreement license           |
| DOCUMENT_RETENTION_TIME              | -        | 604800                                         | Time to live the document in the S3 and DB storage                      |

## Contributing
We feel that a welcoming community is important and we ask that you follow PaySuper's [Open Source Code of Conduct](https://github.com/paysuper/code-of-conduct/blob/master/README.md) in all interactions with the community.

PaySuper welcomes contributions from anyone and everyone. Please refer to each project's style and contribution guidelines for submitting patches and additions. In general, we follow the "fork-and-pull" Git workflow.

The master branch of this repository contains the latest stable release of this component.