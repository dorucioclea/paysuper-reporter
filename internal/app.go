package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/carlescere/goback"
	"github.com/globalsign/mgo"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client/selector/static"
	"github.com/nats-io/stan.go"
	"github.com/nats-io/stan.go/pb"
	pkgBilling "github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	mongodb "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const (
	loggerName = "PAYSUPER_BILLING_REPORTER"
	fileMask   = "report_%s.%s"
)

type Application struct {
	cfg               *config.Config
	log               *zap.Logger
	database          *mongodb.Source
	messageBroker     MessageBrokerInterface
	s3                S3ClientInterface
	documentGenerator DocumentGeneratorInterface
	backOff           goback.SimpleBackoff
	billingService    grpc.BillingService

	fatalFn func(msg string, fields ...zap.Field)
}

func NewApplication() *Application {
	app := &Application{}
	app.initLogger()
	app.initConfig()
	app.initDatabase()
	app.initS3()
	app.initDocumentGenerator()
	app.initBillingServer()

	return app
}

func (app *Application) initLogger() {
	logger, err := zap.NewProduction()

	if err != nil {
		log.Fatalf("Logger initialization failed with error: %s\n", err)
	}

	app.log = logger.Named(loggerName)
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

	app.database, err = mongodb.NewDatabase(mongodb.Mode(mgo.Secondary))

	if err != nil {
		app.fatalFn("Database connection failed", zap.Error(err))
	}

	zap.L().Info("Database initialization successfully...")
}

func (app *Application) initS3() {
	var err error

	app.s3, err = newS3Client(&app.cfg.S3)
	if err != nil {
		app.fatalFn("S3 initialization failed", zap.Error(err))
	}

	zap.L().Info("S3 initialization successfully...")
}

func (app *Application) initDocumentGenerator() {
	var err error

	app.documentGenerator, err = newDocumentGenerator(&app.cfg.DG)
	if err != nil {
		app.fatalFn("Document generator initialization failed", zap.Error(err))
	}

	zap.L().Info("Document generator initialization successfully...")
}

func (app *Application) initBillingServer() {
	options := []micro.Option{
		micro.Name("p1payapi"),
		micro.Version(constant.PayOneMicroserviceVersion),
	}

	if os.Getenv("MICRO_SELECTOR") == "static" {
		log.Println("Use micro selector `static`")
		options = append(options, micro.Selector(static.NewSelector()))
	}

	service := micro.NewService(options...)
	service.Init()

	app.billingService = grpc.NewBillingService(pkgBilling.ServiceName, service.Client())
}

func (app *Application) Run() {
	b := app.backOff
	cb := &b
	for {
		var err error

		ctxStan, cancel := context.WithCancel(context.Background())
		app.messageBroker, err = newMessageBroker(&app.cfg.Nats, cancel)

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

		if err = app.handler(ctxStan); err != nil {
			zap.L().Error("handler error: %v", zap.Error(err))
			goto nextAttempt
		}

		zap.L().Debug("connected to NATS Streaming server, waiting of signal")
		// graceful shutdown
		select {
		case <-ctxStan.Done():
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

	_, err := app.messageBroker.QueueSubscribe(pkg.SubjectRequestReportFileCreate, app.execute, startOpt)
	if err != nil {
		app.messageBroker.Close()
		zap.L().Fatal("Unable to subscribe to the broker message", zap.Error(err))
	}

	return err
}

func (app *Application) execute(msg *stan.Msg) {
	req := &pkg.ReportRequest{}
	if err := json.Unmarshal(msg.Data, req); err != nil {
		zap.L().Error("Invalid message data", zap.Error(err))
		return
	}

	report, err := app.buildReport(req)
	if err != nil {
		zap.L().Error("Unable to build report", zap.Error(err))
		return
	}

	data, err := json.Marshal(report)
	if err != nil {
		zap.L().Error("Unable to marshal report", zap.Error(err))
		return
	}

	payload := &pkg.Payload{
		Template: &pkg.PayloadTemplate{
			ShortId: req.TemplateId,
		},
		Data: data,
	}

	file, err := app.documentGenerator.Render(payload)
	if err != nil {
		zap.L().Error("Unable to render report", zap.Error(err))
		return
	}

	fileName := fmt.Sprintf(fileMask, req.FileId, req.OutputFormat)
	filePath := os.TempDir() + string(os.PathSeparator) + fileName

	if err = ioutil.WriteFile(filePath, file.File, 0644); err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return
	}

	_, err = app.s3.Put(fileName, filePath, PutObjectOptions{})
	if err != nil {
		zap.L().Error("Unable to upload report to the S3", zap.Error(err))
		return
	}

	_, err = app.billingService.UpdateReportFile(
		context.Background(),
		&grpc.UpdateReportFileRequest{Id: req.FileId, FilePath: filePath},
	)
	if err != nil {
		zap.L().Error("Unable to update report", zap.Error(err))
		return
	}
}

func (app *Application) buildReport(req *pkg.ReportRequest) (interface{}, error) {
	var data interface{}
	if err := app.database.Collection(req.TableName).Find(req.Match).All(&data); err != nil {
		return nil, err
	}

	return data, nil
}
