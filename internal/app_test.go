package internal

import (
	"github.com/paysuper/paysuper-reporter/internal/mock"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"os"
	"testing"
)

type ApplicationTestSuite struct {
	suite.Suite
	app         *Application
	zapRecorder *observer.ObservedLogs
}

func Test_Application(t *testing.T) {
	suite.Run(t, new(ApplicationTestSuite))
}

func (suite *ApplicationTestSuite) SetupTest() {
	suite.app = NewApplication()
	assert.NotNil(suite.T(), suite.app)
	assert.NotNil(suite.T(), suite.app.cfg)
	assert.NotNil(suite.T(), suite.app.log)
	assert.NotNil(suite.T(), suite.app.broker)
	assert.NotNil(suite.T(), suite.app.httpClient)

	suite.app.broker = mock.NewBrokerMockOk()
	suite.app.httpClient = mock.NewClientStatusOk()

	var core zapcore.Core

	lvl := zap.NewAtomicLevel()
	core, suite.zapRecorder = observer.New(lvl)
	suite.app.log = zap.New(core)
	zap.ReplaceGlobals(suite.app.log)
}

func (suite *ApplicationTestSuite) TearDownTest() {}

func (suite *ApplicationTestSuite) TestApplication_InitBroker_Ok() {
	suite.app.initBroker()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Broker created...", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_InitBroker_NewBroker_Error() {
	suite.app.fatalFn = zap.L().Info
	suite.app.cfg.BrokerAddress = ""
	suite.app.broker = nil

	suite.app.initBroker()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Creating RabbitMq broker failed", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_InitBroker_RegisterSubscriber_Error() {
	suite.app.fatalFn = zap.L().Info
	suite.app.cfg.BrokerAddress = ""
	suite.app.broker = mock.NewBrokerMockError()

	suite.app.initBroker()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Registration RabbitMQ broker handler failed", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_InitConfig_Ok() {
	suite.app.initConfig()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Configuration parsed successfully...", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_InitConfig_Error() {
	v := os.Getenv("POSTMARK_EMAIL_FROM")
	err := os.Unsetenv("POSTMARK_EMAIL_FROM")
	assert.NoError(suite.T(), err)

	suite.app.fatalFn = zap.L().Info

	suite.app.initConfig()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Config init failed", messages[0].Message)

	err = os.Setenv("POSTMARK_EMAIL_FROM", v)
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationTestSuite) TestApplication_Run_Ok() {
	suite.app.Run()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Application started...", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_Run_Subscribe_Error() {
	suite.app.broker = mock.NewBrokerMockError()
	suite.app.fatalFn = zap.L().Info

	suite.app.Run()
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Application subscriber start failed...", messages[1].Message)
}

func (suite *ApplicationTestSuite) TestApplication_EmailConfirmProcess_Ok() {
	payload := &proto.Payload{
		TemplateAlias: "template1",
		TemplateModel: map[string]string{"param1": "value1"},
	}

	err := suite.app.emailProcess(payload, amqp.Delivery{})
	assert.NoError(suite.T(), err)
}

func (suite *ApplicationTestSuite) TestApplication_EmailConfirmProcess_CreatingHttpRequestError() {
	payload := &proto.Payload{
		TemplateAlias: "template1",
		TemplateModel: map[string]string{"param1": "value1"},
	}

	suite.app.cfg.PostmarkApiUrl = "\n"

	err := suite.app.emailProcess(payload, amqp.Delivery{})
	assert.Error(suite.T(), err)
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Creating http request failed", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_EmailConfirmProcess_SendEmailFailedError() {
	payload := &proto.Payload{
		TemplateAlias: "template1",
		TemplateModel: map[string]string{"param1": "value1"},
	}

	suite.app.httpClient = mock.NewTransportHttpError()

	err := suite.app.emailProcess(payload, amqp.Delivery{})
	assert.Error(suite.T(), err)
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Send email failed", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_EmailConfirmProcess_ReadingResponseBodyError() {
	payload := &proto.Payload{
		TemplateAlias: "template1",
		TemplateModel: map[string]string{"param1": "value1"},
	}

	suite.app.httpClient = mock.NewClientStatusErrorIoReader()

	err := suite.app.emailProcess(payload, amqp.Delivery{})
	assert.Error(suite.T(), err)
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Reading response body failed", messages[0].Message)
}

func (suite *ApplicationTestSuite) TestApplication_EmailConfirmProcess_IncorrectJsonResponseError() {
	payload := &proto.Payload{
		TemplateAlias: "template1",
		TemplateModel: map[string]string{"param1": "value1"},
	}
	suite.app.httpClient = mock.NewClientStatusIncorrectResponse()

	err := suite.app.emailProcess(payload, amqp.Delivery{})
	assert.Error(suite.T(), err)
	messages := suite.zapRecorder.All()
	assert.Equal(suite.T(), "Incorrect json response", messages[0].Message)
}
