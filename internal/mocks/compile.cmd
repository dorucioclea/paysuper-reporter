mockery -recursive=true -all -dir=../repository -output=.
mockery -recursive=true -name=CentrifugoInterface -dir=../ -output=.
mockery -recursive=true -name=DocumentGeneratorInterface -dir=../ -output=.
mockery -name=ReporterService -dir=../../pkg/proto -recursive=true -output=../../pkg/mocks

