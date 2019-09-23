package internal

import (
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CentrifugoTestSuite struct {
	suite.Suite
}

func Test_Centrifugo(t *testing.T) {
	suite.Run(t, new(CentrifugoTestSuite))
}

func (suite *CentrifugoTestSuite) TestCentrifugo_newCentrifugoClient_Ok() {
	centrifugo := newCentrifugoClient(&config.CentrifugoConfig{})
	assert.IsType(suite.T(), &Centrifugo{}, centrifugo)
}

func (suite *CentrifugoTestSuite) TestCentrifugo_Publish_Error_Marshal() {
	centrifugo := newCentrifugoClient(&config.CentrifugoConfig{})
	assert.Error(suite.T(), centrifugo.Publish("string", make(chan int)))
}

func (suite *CentrifugoTestSuite) TestCentrifugo_Publish_Error_Client() {
	centrifugo := newCentrifugoClient(&config.CentrifugoConfig{})
	assert.Error(suite.T(), centrifugo.Publish("string", "test"))
}
