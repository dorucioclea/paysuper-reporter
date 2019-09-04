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
	"time"
)

type Application struct {
	cfg                    *config.Config
	log                    *zap.Logger
	database               *mongodb.Source
	messageBroker          nats.NatsManagerInterface
	s3                     awsWrapper.AwsManagerInterface
	centrifugo             CentrifugoInterface
	documentGenerator      DocumentGeneratorInterface
	reportFileRepository   repository.ReportFileRepositoryInterface
	royaltyRepository      repository.RoyaltyRepositoryInterface
	vatRepository          repository.VatRepositoryInterface
	transactionsRepository repository.TransactionsRepositoryInterface

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

	app.reportFileRepository = repository.NewReportFileRepository(app.database)
	app.royaltyRepository = repository.NewRoyaltyReportRepository(app.database)
	app.vatRepository = repository.NewVatRepository(app.database)
	app.transactionsRepository = repository.NewTransactionsRepository(app.database)

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
		app.fatalFn("S3 initialization failed", zap.Error(err))
	}

	zap.L().Info("S3 initialization successfully...")
}

func (app *Application) initCentrifugo() {
	app.centrifugo = newCentrifugoClient(&app.cfg.CentrifugoConfig)

	zap.L().Info("Centrifugo initialization successfully...")
}

func (app *Application) initDocumentGenerator() {
	var err error

	app.documentGenerator, err = newDocumentGenerator(&app.cfg.DG)
	if err != nil {
		app.fatalFn("Document generator initialization failed", zap.Error(err))
	}

	zap.L().Info("Document generator initialization successfully...")
}

func (app *Application) initMessageBroker() {
	var err error

	app.messageBroker, err = nats.NewNatsManager()

	if err != nil {
		app.fatalFn("Message broker initialization failed", zap.Error(err))
	}

	zap.L().Info("Message broker initialization successfully...")
}

func (app *Application) Run() {
	var service micro.Service
	options := []micro.Option{
		micro.Name(pkg.ServiceName),
		micro.Version(pkg.ServiceVersion),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
	}

	if app.cfg.MicroSelector == "static" {
		zap.L().Info(`Use micro selector "static"`)
		options = append(options, micro.Selector(static.NewSelector()))
	}

	service = micro.NewService(options...)
	service.Init()

	if err := proto.RegisterReporterServiceHandler(service.Server(), app); err != nil {
		app.fatalFn("Can`t register service in micro", zap.Error(err))
	}

	if err := service.Run(); err != nil {
		app.fatalFn("Can`t run service", zap.Error(err))
	}

	startOpt := stan.StartAt(pb.StartPosition_NewOnly)
	_, err := app.messageBroker.QueueSubscribe(pkg.SubjectRequestReportFileCreate, "", app.execute, startOpt)

	if err != nil {
		zap.L().Fatal("Unable to subscribe to the broker message", zap.Error(err))
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
	reportFile := &proto.MgoReportFile{}

	if err := json.Unmarshal(msg.Data, reportFile); err != nil {
		zap.L().Error("Invalid message data", zap.Error(err))
		return
	}

	h := builder.NewBuilder(
		reportFile,
		app.reportFileRepository,
		app.royaltyRepository,
		app.vatRepository,
		app.transactionsRepository,
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

	data, err := json.Marshal(rawData)

	if err != nil {
		zap.L().Error("Unable to marshal report", zap.Error(err))
		return
	}

	payload := &proto.GeneratorPayload{
		Template: &proto.GeneratorTemplate{
			ShortId: reportFile.Template,
		},
		Data: data,
	}

	file, err := app.documentGenerator.Render(payload)
	if err != nil {
		zap.L().Error("Unable to render report", zap.Error(err))
		return
	}

	fileName := fmt.Sprintf(pkg.FileMask, reportFile.Id, reportFile.FileType)
	filePath := os.TempDir() + string(os.PathSeparator) + fileName

	if err = ioutil.WriteFile(filePath, file.File, 0644); err != nil {
		zap.L().Error("internal error", zap.Error(err))
		return
	}

	_, err = app.s3.Upload(context.TODO(), &awsWrapper.UploadInput{Body: bytes.NewReader(file.File), FileName: fileName})

	if err != nil {
		zap.L().Error("Unable to upload report to the S3", zap.Error(err))
		return
	}

	err = app.centrifugo.Publish(fmt.Sprintf(app.cfg.CentrifugoConfig.MerchantChannel, reportFile.MerchantId), file)

	if err != nil {
		zap.L().Error(
			errors.ErrorCentrifugoNotificationFailed.Message,
			zap.Error(err),
			zap.Any("report_file", reportFile),
		)
		return
	}

	return
}

func (c *appHealthCheck) Status() (interface{}, error) {
	// INFO: Always is fail on locally if your DB don't have secondary members of the replica set
	if err := c.db.Ping(); err != nil {
		return "fail", err
	}

	return "ok", nil
}
