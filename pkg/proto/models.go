package proto

import (
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"time"
)

type MgoReportFile struct {
	Id         bson.ObjectId          `bson:"_id"`
	MerchantId bson.ObjectId          `bson:"merchant_id"`
	FileType   string                 `bson:"file_type"`
	ReportType string                 `bson:"report_type"`
	Template   string                 `bson:"template"`
	Params     map[string]interface{} `bson:"params"`
	CreatedAt  time.Time              `bson:"created_at"`
}

func (m *ReportFile) GetBSON() (interface{}, error) {
	st := &MgoReportFile{
		Id:         bson.ObjectIdHex(m.Id),
		MerchantId: bson.ObjectIdHex(m.MerchantId),
		Template:   m.Template,
		FileType:   m.FileType,
		ReportType: m.ReportType,
		CreatedAt:  time.Now(),
	}

	if m.Params != nil {
		if err := json.Unmarshal(m.Params, st.Params); err != nil {
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

func (m *ReportFile) SetBSON(raw bson.Raw) error {
	decoded := new(MgoReportFile)
	err := raw.Unmarshal(decoded)

	if err != nil {
		return err
	}

	m.Id = decoded.Id.Hex()
	m.MerchantId = decoded.MerchantId.Hex()
	m.Template = decoded.Template
	m.ReportType = decoded.ReportType
	m.FileType = decoded.FileType

	if decoded.Params != nil {
		if m.Params, err = json.Marshal(decoded.Params); err != nil {
			return err
		}
	}

	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}
