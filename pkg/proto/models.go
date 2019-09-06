package proto

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"time"
)

type GeneratorPayload struct {
	Template *GeneratorTemplate `json:"template"`
	Options  *GeneratorOptions  `json:"options"`
	Data     interface{}        `json:"data"`
}

type GeneratorTemplate struct {
	ShortId string `json:"shortid,omitempty"`
	Name    string `json:"name,omitempty"`
	Recipe  string `json:"recipe,omitempty"`
	Content string `json:"content,omitempty"`
}

type GeneratorOptions struct {
	Timeout string `json:"timeout,omitempty"`
}

type MgoReportFile struct {
	Id         bson.ObjectId          `bson:"_id"`
	MerchantId bson.ObjectId          `bson:"merchant_id"`
	FileType   string                 `bson:"file_type"`
	ReportType string                 `bson:"report_type"`
	TemplateId string                 `bson:"template_id"`
	Params     map[string]interface{} `bson:"params"`
	CreatedAt  time.Time              `bson:"created_at"`
	ExpireAt   time.Time              `bson:"expire_at"`
}

func (m *ReportFile) GetBSON() (interface{}, error) {
	st := &MgoReportFile{
		Id:         bson.ObjectIdHex(m.Id),
		MerchantId: bson.ObjectIdHex(m.MerchantId),
		TemplateId: m.Template,
		FileType:   m.FileType,
		ReportType: m.ReportType,
		CreatedAt:  time.Now(),
	}

	if m.Params != nil {
		if err := json.Unmarshal(m.Params, &st.Params); err != nil {
			return nil, err
		}
	}

	if m.CreatedAt != nil {
		t, err := ptypes.Timestamp(m.CreatedAt)
		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	}

	return st, nil
}
