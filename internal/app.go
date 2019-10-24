package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
	nats "github.com/ProtocolONE/nats/pkg"
	"github.com/micro/go-micro"
	"github.com/micro/go-plugins/client/selector/static"
	"github.com/micro/go-plugins/wrapper/monitoring/prometheus"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/internal/builder"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/internal/repository"
	"github.com/paysuper/paysuper-reporter/pkg"
	"github.com/paysuper/paysuper-reporter/pkg/errors"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Application struct {
	cfg                    *config.Config
	log                    *zap.Logger
	database               *mongodb.Source
	messageBroker          nats.NatsManagerInterface
	s3                     awsWrapper.AwsManagerInterface
	s3Agreement            awsWrapper.AwsManagerInterface
	centrifugo             CentrifugoInterface
	documentGenerator      DocumentGeneratorInterface
	royaltyRepository      repository.RoyaltyRepositoryInterface
	vatRepository          repository.VatRepositoryInterface
	transactionsRepository repository.TransactionsRepositoryInterface
	payoutRepository       repository.PayoutRepositoryInterface
	merchantRepository     repository.MerchantRepositoryInterface
	service                micro.Service

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

	app.royaltyRepository = repository.NewRoyaltyReportRepository(app.database)
	app.vatRepository = repository.NewVatRepository(app.database)
	app.transactionsRepository = repository.NewTransactionsRepository(app.database)
	app.payoutRepository = repository.NewPayoutRepository(app.database)
	app.merchantRepository = repository.NewMerchantRepository(app.database)

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
	var err error

	opts := []nats.Option{
		nats.ClientId(app.cfg.Nats.ClientId + "_" + strconv.FormatInt(time.Now().UnixNano(), 16)),
	}
	app.messageBroker, err = nats.NewNatsManager(opts...)

	if err != nil {
		app.fatalFn("Message broker initialization failed", zap.Error(err))
	}

	zap.L().Info("Message broker initialization successfully...")
}

func (app *Application) Run() {
	options := []micro.Option{
		micro.Name(pkg.ServiceName),
		micro.Version(pkg.ServiceVersion),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
		micro.BeforeStart(func() error {
			go func() {
				startOpt := stan.StartAt(pb.StartPosition_NewOnly)
				_, err := app.messageBroker.QueueSubscribe(pkg.SubjectRequestReportFileCreate, "", app.execute, startOpt)

				if err != nil {
					zap.L().Fatal("Unable to subscribe to the broker message", zap.Error(err))
				}
			}()

			return nil
		}),
		micro.AfterStop(func() error {
			if err := app.messageBroker.Close(); err != nil {
				zap.L().Fatal("Unable to close the broker message", zap.Error(err))
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

	if err := proto.RegisterReporterServiceHandler(app.service.Server(), app); err != nil {
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

func (app *Application) execute(msg *stan.Msg) {
	reportFile := &proto.ReportFile{}

	if err := json.Unmarshal(msg.Data, reportFile); err != nil {
		zap.L().Error("Invalid message data", zap.Error(err))
		return
	}

	h := builder.NewBuilder(
		app.service,
		reportFile,
		app.royaltyRepository,
		app.vatRepository,
		app.transactionsRepository,
		app.payoutRepository,
		app.merchantRepository,
	)
	bldr, err := h.GetBuilder()

	if err != nil {
		zap.L().Error("Unable to get builder", zap.Error(err))
		return
	}

	rawData, err := bldr.Build()

	if err != nil {
		zap.L().Error("Unable to build document", zap.Error(err))
		return
	}

	payload := &proto.GeneratorPayload{
		Template: &proto.GeneratorTemplate{
			ShortId: reportFile.Template,
			Recipe:  reportFileRecipes[reportFile.FileType],
		},
		Data: rawData,
	}

	file, err := app.documentGenerator.Render(payload)
	if err != nil {
		zap.L().Error("Unable to render report", zap.Error(err), zap.Any("payload", payload))
		return
	}

	fileName := fmt.Sprintf(pkg.FileMask, reportFile.UserId, reportFile.Id, reportFile.FileType)

	if reportFile.ReportType == pkg.ReportTypeAgreement {
		fileName = fmt.Sprintf(pkg.FileMaskAgreement, reportFile.MerchantId, reportFile.FileType)
	}

	filePath := os.TempDir() + string(os.PathSeparator) + fileName

	if err = ioutil.WriteFile(filePath, file, 0644); err != nil {
		zap.L().Error("internal error", zap.Error(err))
		return
	}

	retentionTime := app.cfg.DocumentRetentionTime
	if reportFile.RetentionTime > 0 {
		retentionTime = int(reportFile.RetentionTime)
	}

	ctx := context.TODO()
	awsManager := app.s3
	in := &awsWrapper.UploadInput{Body: bytes.NewReader(file), FileName: fileName}

	if reportFile.ReportType == pkg.ReportTypeAgreement {
		awsManager = app.s3Agreement
	} else {
		in.Expires = time.Now().Add(time.Duration(retentionTime) * time.Second)
	}

	_, err = awsManager.Upload(ctx, in)

	if err != nil {
		zap.L().Error("Unable to upload report to the S3", zap.Error(err))
		return
	}

	if reportFile.SendNotification {
		msg := map[string]string{"file_name": reportFile.Id + "." + reportFile.FileType}
		err = app.centrifugo.Publish(fmt.Sprintf(app.cfg.CentrifugoConfig.UserChannel, reportFile.MerchantId), msg)

		if err != nil {
			zap.L().Error(
				errors.ErrorCentrifugoNotificationFailed.Message,
				zap.Error(err),
				zap.Any("report_file", reportFile),
			)
			return
		}
	}

	if err = os.Remove(filePath); err != nil {
		zap.L().Error(
			"Unable to delete temporary file",
			zap.Error(err),
			zap.String("path", filePath),
		)
		return
	}

	if err = bldr.PostProcess(ctx, reportFile.Id, fileName, retentionTime); err != nil {
		zap.L().Error(
			"PostProcess execution error",
			zap.Error(err),
			zap.String("path", filePath),
		)
		return
	}

	return
}

func (c *appHealthCheck) Status() (interface{}, error) {
	// INFO: Always is fail on locally if your DB don't have secondary members of the replica set
	// and use secondary mode of database connection
	if err := c.db.Ping(); err != nil {
		return "fail", err
	}

	return "ok", nil
}
