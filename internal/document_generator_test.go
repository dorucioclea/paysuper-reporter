package internal

import (
	"github.com/paysuper/paysuper-reporter/internal/config"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DocumentGeneratorTestSuite struct {
	suite.Suite
}

func Test_DocumentGenerator(t *testing.T) {
	suite.Run(t, new(DocumentGeneratorTestSuite))
}

func (suite *DocumentGeneratorTestSuite) TestDocumentGenerator_newDocumentGenerator_Ok() {
	dg := newDocumentGenerator(&config.DocumentGeneratorConfig{})
	assert.IsType(suite.T(), &DocumentGenerator{}, dg)
}

func (suite *CentrifugoTestSuite) TestDocumentGenerator_Render_Error_Marshal() {
	dg := newDocumentGenerator(&config.DocumentGeneratorConfig{})
	_, err := dg.Render(&proto.GeneratorPayload{Data: make(chan int)})
	assert.Error(suite.T(), err)
}

func (suite *CentrifugoTestSuite) TestDocumentGenerator_Render_Error_Client() {
	dg := newDocumentGenerator(&config.DocumentGeneratorConfig{})
	_, err := dg.Render(&proto.GeneratorPayload{})
	assert.Error(suite.T(), err)
}
