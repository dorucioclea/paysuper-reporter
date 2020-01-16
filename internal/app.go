package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
	protobufProto "github.com/golang/protobuf/proto"
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/client/selector/static"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-proto/go/billingpb"
	"github.com/paysuper/paysuper-proto/go/reporterpb"
	"github.com/paysuper/paysuper-reporter/internal/builder"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	rabbitmq "gopkg.in/ProtocolONE/rabbitmq.v1/pkg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Application struct {
	cfg               *config.Config
	log               *zap.Logger
	database          *mongodb.Source
	s3                awsWrapper.AwsManagerInterface
	s3Agreement       awsWrapper.AwsManagerInterface
	centrifugo        CentrifugoInterface
	documentGenerator DocumentGeneratorInterface
	service           micro.Service
	billing           billingpb.BillingService

	generateReportBroker rabbitmq.BrokerInterface
	postProcessBroker    rabbitmq.BrokerInterface

	fatalFn func(msg string, fields ...zap.Field)
}

type appHealthCheck struct {
	db *mongodb.Source
}

func NewApplication() *Application {
	app := &Application{}
	app.initLogger()
	app.initConfig()
	app.initDatabase()
	app.initS3()
	app.initCentrifugo()
	app.initDocumentGenerator()
	app.initMessageBroker()
	app.initHealth()

	return app
}

func (app *Application) initHealth() {
	h := health.New()
	err := h.AddChecks([]*health.Config{
		{
			Name: "health-check",
			Checker: &appHealthCheck{
				db: app.database,
			},
			Interval: time.Duration(1) * time.Second,
			Fatal:    true,
		},
	})

	if err != nil {
		app.fatalFn("Health check register failed", zap.Error(err))
	}

	if err = h.Start(); err != nil {
		app.fatalFn("Health check start failed", zap.Error(err))
	}

	app.log.Info("Health check listener started", zap.String("port", app.cfg.MetricsPort))

	http.HandleFunc("/health", handlers.NewJSONHandlerFunc(h, nil))
}

func (app *Application) initLogger() {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("Logger initialization failed with error: %s\n", err)
	}

	app.log = logger.Named(pkg.LoggerName)
	zap.ReplaceGlobals(app.log)

	app.fatalFn = zap.L().Fatal

	zap.L().Info("Logger init...")
}

func (app *Application) initConfig() {
	var err error

	app.cfg, err = config.NewConfig()

	if err != nil {
		app.fatalFn("Config init failed", zap.Error(err))
	}

	zap.L().Info("Configuration parsed successfully...")
}

func (app *Application) initDatabase() {
	var err error

	app.database, err = mongodb.NewDatabase(mongodb.Mode(app.cfg.Db.MongoMode))

	if err != nil {
		app.fatalFn("Database connection failed", zap.Error(err))
	}

	zap.L().Info("Database initialization successfully...")
}

func (app *Application) initS3() {
	var err error

	app.s3, err = awsWrapper.New()
	if err != nil {
		app.fatalFn("reports S3 initialization failed", zap.Error(err))
	}

	zap.L().Info("reports S3 initialization successfully...")

	awsOptions := []awsWrapper.Option{
		awsWrapper.AccessKeyId(app.cfg.S3.AwsAccessKeyIdAgreement),
		awsWrapper.SecretAccessKey(app.cfg.S3.AwsSecretAccessKeyAgreement),
		awsWrapper.Region(app.cfg.S3.AwsRegionAgreement),
		awsWrapper.Bucket(app.cfg.S3.AwsBucketAgreement),
	}
	app.s3Agreement, err = awsWrapper.New(awsOptions...)

	if err != nil {
		app.fatalFn("agreement S3 initialization failed", zap.Error(err))
	}

	zap.L().Info("agreement S3 initialization successfully...")
}

func (app *Application) initCentrifugo() {
	app.centrifugo = newCentrifugoClient(&app.cfg.CentrifugoConfig)

	zap.L().Info("Centrifugo initialization successfully...")
}

func (app *Application) initDocumentGenerator() {
	app.documentGenerator = newDocumentGenerator(&app.cfg.DG)

	zap.L().Info("Document generator initialization successfully...")
}

func (app *Application) initMessageBroker() {
	generateReportBroker, err := rabbitmq.NewBroker(app.cfg.BrokerAddress)

	if err != nil {
		app.fatalFn(
			"Creating generate report broker failed",
			zap.Error(err),
			zap.String("DSN", app.cfg.BrokerAddress),
		)
		return
	}

	generateReportBroker.SetExchangeName(pkg.BrokerGenerateReportTopicName)
	err = generateReportBroker.RegisterSubscriber(pkg.BrokerGenerateReportTopicName, app.ExecuteProcess)

	if err != nil {
		app.fatalFn("Registration generate report subscriber function failed", zap.Error(err))
		return
	}

	postProcessBroker, err := rabbitmq.NewBroker(app.cfg.BrokerAddress)

	if err != nil {
		app.fatalFn(
			"Creating post process broker failed",
			zap.Error(err),
			zap.String("DSN", app.cfg.BrokerAddress),
		)
		return
	}

	postProcessBroker.SetExchangeName(pkg.BrokerPostProcessTopicName)
	err = postProcessBroker.RegisterSubscriber(pkg.BrokerPostProcessTopicName, app.ExecutePostProcess)

	if err != nil {
		app.fatalFn("Registration post process subscriber function failed", zap.Error(err))
		return
	}

	app.generateReportBroker = generateReportBroker
	app.postProcessBroker = postProcessBroker

	zap.L().Info("Message brokers initialized successfully...")
}

func (app *Application) Run() {
	options := []micro.Option{
		micro.Name(pkg.ServiceName),
		micro.Version(pkg.ServiceVersion),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.BeforeStart(func() error {
			go func() {
				go func() {
					err := app.generateReportBroker.Subscribe(nil)

					if err != nil {
						app.fatalFn("Generate report subscriber start failed...", zap.Error(err))
					}
				}()

				err := app.postProcessBroker.Subscribe(nil)

				if err != nil {
					app.fatalFn("Generate report subscriber start failed...", zap.Error(err))
				}
			}()

			return nil
		}),
		micro.AfterStop(func() error {
			if err := app.log.Sync(); err != nil {
				app.fatalFn("Logger sync failed", zap.Error(err))
			} else {
				zap.L().Info("Logger synced")
			}

			return nil
		}),
	}

	if app.cfg.MicroSelector == "static" {
		zap.L().Info(`Use micro selector "static"`)
		options = append(options, micro.Selector(static.NewSelector()))
	}

	app.service = micro.NewService(options...)
	app.service.Init()

	app.billing = billingpb.NewBillingService(billingpb.ServiceName, app.service.Client())

	if err := reporterpb.RegisterReporterServiceHandler(app.service.Server(), app); err != nil {
		app.fatalFn("Can`t register service in micro", zap.Error(err))
	}

	if err := app.service.Run(); err != nil {
		app.fatalFn("Can`t run service", zap.Error(err))
	}
}

func (app *Application) Stop() {
	if err := app.log.Sync(); err != nil {
		app.fatalFn("Logger sync failed", zap.Error(err))
	} else {
		zap.L().Info("Logger synced")
	}
}

func (app *Application) ExecuteProcess(payload *reporterpb.ReportFile, d amqp.Delivery) error {
	h := builder.NewBuilder(
		app.service,
		payload,
		app.billing,
	)
	handler, err := h.GetBuilder()

	if err != nil {
		zap.L().Error(
			"Unable to get handler",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	rawData, err := handler.Build()

	if err != nil {
		zap.L().Error(
			"Unable to build document",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	fileRequest := &proto.GeneratorPayload{
		Template: &proto.GeneratorTemplate{
			ShortId: payload.Template,
			Recipe:  reportFileRecipes[payload.FileType],
		},
		Data: rawData,
	}

	file, err := app.documentGenerator.Render(fileRequest)

	if err != nil {
		zap.L().Error(
			"Unable to render report",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	fileName := fmt.Sprintf(pkg.FileMask, payload.UserId, payload.Id, payload.FileType)

	if payload.ReportType == pkg.ReportTypeAgreement {
		fileName = fmt.Sprintf(pkg.FileMaskAgreement, payload.MerchantId, payload.FileType)
	}

	filePath := os.TempDir() + string(os.PathSeparator) + fileName
	err = ioutil.WriteFile(filePath, file, 0644)

	if err != nil {
		zap.L().Error(
			"internal error",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	retentionTime := app.cfg.DocumentRetentionTime

	if payload.RetentionTime > 0 {
		retentionTime = int64(payload.RetentionTime)
	}

	awsManager := app.s3
	in := &awsWrapper.UploadInput{
		Body:     bytes.NewReader(file),
		FileName: fileName,
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	if payload.ReportType == pkg.ReportTypeAgreement {
		awsManager = app.s3Agreement
	} else {
		in.Expires = time.Now().Add(time.Duration(retentionTime) * time.Second)
	}

	_, err = awsManager.Upload(ctx, in)

	if err != nil {
		zap.L().Error(
			"Unable to upload report to the S3",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	if payload.SendNotification {
		msg := map[string]string{"file_name": payload.Id + "." + payload.FileType}
		ch := fmt.Sprintf(app.cfg.CentrifugoConfig.UserChannel, payload.MerchantId)
		err = app.centrifugo.Publish(ch, msg)

		if err != nil {
			zap.L().Error(
				errors.ErrorCentrifugoNotificationFailed.Message,
				zap.Error(err),
				zap.Any("payload", payload),
			)
			return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
		}
	}

	err = os.Remove(filePath)

	if err != nil {
		zap.L().Error(
			"Unable to delete temporary file",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	postProcessData := &reporterpb.PostProcessRequest{
		ReportFile:    payload,
		FileName:      fileName,
		RetentionTime: retentionTime,
		File:          file,
	}
	amqpHeaders := amqp.Table{
		"x-retry-count": int32(0),
	}
	err = app.postProcessBroker.Publish(pkg.BrokerPostProcessTopicName, postProcessData, amqpHeaders)

	if err != nil {
		postProcessData.File = nil
		zap.L().Error(
			"Publish message to post process broker failed",
			zap.Error(err),
			zap.Any("data", postProcessData),
		)
		return app.getProcessResult(app.generateReportBroker, pkg.BrokerGenerateReportTopicName, payload, d)
	}

	return nil
}

func (app *Application) ExecutePostProcess(payload *reporterpb.PostProcessRequest, d amqp.Delivery) error {
	log.Println("2")
	h := builder.NewBuilder(
		app.service,
		payload.ReportFile,
		app.billing,
	)
	handler, err := h.GetBuilder()

	if err != nil {
		payload.File = nil
		zap.L().Error(
			"Unable to get handler",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.postProcessBroker, pkg.BrokerPostProcessTopicName, payload, d)
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Minute*2)
	err = handler.PostProcess(ctx, payload.ReportFile.Id, payload.FileName, payload.RetentionTime, payload.File)

	if err != nil {
		payload.File = nil
		zap.L().Error(
			"PostProcess execution error",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		return app.getProcessResult(app.postProcessBroker, pkg.BrokerPostProcessTopicName, payload, d)
	}

	return nil
}

func (app *Application) getProcessResult(
	broker rabbitmq.BrokerInterface,
	topic string,
	message protobufProto.Message,
	d amqp.Delivery,
) error {
	retryCount := int32(0)

	if v, ok := d.Headers[rabbitmq.BrokerMessageRetryCountHeader]; ok {
		retryCount = v.(int32)
	}

	if retryCount >= pkg.BrokerMessageRetryMaxCount {
		return nil
	}

	amqpHeaders := amqp.Table{
		"x-retry-count": retryCount + 1,
	}
	err := broker.Publish(topic, message, amqpHeaders)

	if err != nil {
		zap.L().Error(
			"ReQueue message to broker failed",
			zap.Error(err),
			zap.String("topic", topic),
			zap.Any("message", message),
			zap.Any("headers", amqpHeaders),
		)
		return nil
	}

	return nil
}

func (c *appHealthCheck) Status() (interface{}, error) {
	// INFO: Always is fail on locally if your DB don't have secondary members of the replica set
	// and use secondary mode of database connection
	if err := c.db.Ping(); err != nil {
		return "fail", err
	}

	return "ok", nil
}
