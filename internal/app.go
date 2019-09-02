package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/InVisionApp/go-health"
	"github.com/InVisionApp/go-health/handlers"
	nats "github.com/ProtocolONE/nats/pkg"
	"github.com/carlescere/goback"
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

const (
	fileMask = "report_%s.%s"
)

type Application struct {
	cfg                     *config.Config
	log                     *zap.Logger
	database                *mongodb.Source
	messageBroker           nats.NatsManagerInterface
	s3                      awsWrapper.AwsManagerInterface
	centrifugo              CentrifugoInterface
	documentGenerator       DocumentGeneratorInterface
	backOff                 goback.SimpleBackoff
	router                  *http.ServeMux
	reportFileRepository    repository.ReportFileRepositoryInterface
	royaltyReportRepository repository.RoyaltyReportRepositoryInterface
	vatReportRepository     repository.VatReportRepositoryInterface

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

	app.reportFileRepository = repository.NewReportFileRepository(app.database)
	app.royaltyReportRepository = repository.NewRoyaltyReportRepository(app.database)
	app.vatReportRepository = repository.NewVatReportRepository(app.database)

	app.router = http.NewServeMux()
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

	app.router.HandleFunc("/health", handlers.NewJSONHandlerFunc(h, nil))
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
}

func (app *Application) initDocumentGenerator() {
	var err error

	app.documentGenerator, err = newDocumentGenerator(&app.cfg.DG)
	if err != nil {
		app.fatalFn("Document generator initialization failed", zap.Error(err))
	}

	zap.L().Info("Document generator initialization successfully...")
}

func (app *Application) Run() {
	b := app.backOff
	cb := &b
	for {
		var err error

		ctx := context.TODO()
		app.messageBroker, err = nats.New()

		if err != nil {
			zap.L().Error("connect to NATS Streaming server failed", zap.Error(err))
			// Next attempt
			d, err := cb.NextAttempt()
			if err != nil {
				zap.L().Error("backoff error", zap.Error(err))
			}
			if d < 0 {
				d = 0
			}
			time.Sleep(d)
			continue
		}

		zap.L().Info("Message broker initialization successfully...")

		cb.Reset()

		if err = app.handler(ctx); err != nil {
			zap.L().Error("handler error: %v", zap.Error(err))
			goto nextAttempt
		}

		zap.L().Debug("connected to NATS Streaming server, waiting of signal")
		// graceful shutdown
		select {
		case <-ctx.Done():
		}

	nextAttempt:
		if err := app.messageBroker.Close(); err != nil {
			zap.L().Error("Close connection to NATS Streaming server failed", zap.Error(err))
		}

		zap.L().Debug("retry to reconnect to NATS Streaming server")
		d, err := cb.NextAttempt()
		if err != nil {
			zap.L().Error("BackOff attempt error", zap.Error(err))
		}
		// fix bug with negative time
		if d < 0 {
			cb.Reset()
			d = 0
		}
		time.Sleep(d)
		continue
	}
}

func (app *Application) Stop() {
	if err := app.log.Sync(); err != nil {
		app.fatalFn("Logger sync failed", zap.Error(err))
	} else {
		zap.L().Info("Logger synced")
	}
}

func (app *Application) handler(ctx context.Context) error {
	startOpt := stan.StartAt(pb.StartPosition_NewOnly)

	_, err := app.messageBroker.QueueSubscribe(pkg.SubjectRequestReportFileCreate, "", app.execute, startOpt)
	if err != nil {
		app.messageBroker.Close()
		zap.L().Fatal("Unable to subscribe to the broker message", zap.Error(err))
	}

	return err
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
		app.royaltyReportRepository,
		app.vatReportRepository,
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

	fileName := fmt.Sprintf(fileMask, reportFile.Id, reportFile.FileType)
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
	if err := c.db.Ping(); err != nil {
		return "fail", err
	}
	return "ok", nil
}
