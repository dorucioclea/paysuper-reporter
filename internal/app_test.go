package internal

import (
	"github.com/nats-io/stan.go"
	"github.com/stretchr/testify/suite"
	"testing"
)

type AppTestSuite struct {
	suite.Suite
}

func Test_App(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (suite *AppTestSuite) TestApp_execute_Error_UnmarshalMessage() {
	app := &Application{}
	app.execute(
		&stan.Msg{Data: []byte{}},
	)
}
