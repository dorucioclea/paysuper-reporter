package proto

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"time"
)

type MgoReportFile struct {
	Id         bson.ObjectId `bson:"_id"`
	MerchantId bson.ObjectId `bson:"merchant_id"`
	Type       string        `bson:"type"`
	DateFrom   time.Time     `bson:"date_from"`
	DateTo     time.Time     `bson:"date_to"`
	CreatedAt  time.Time     `bson:"created_at"`
}

func (m *ReportFile) GetBSON() (interface{}, error) {
	st := &MgoReportFile{
		Id:         bson.ObjectIdHex(m.Id),
		MerchantId: bson.ObjectIdHex(m.MerchantId),
		Type:       m.Type,
		CreatedAt:  time.Now(),
	}

	if m.CreatedAt != nil {
		t, err := ptypes.Timestamp(m.CreatedAt)
		if err != nil {
			return nil, err
		}

		st.CreatedAt = t
	}

	if m.DateFrom != nil {
		t, err := ptypes.Timestamp(m.DateFrom)
		if err != nil {
			return nil, err
		}

		st.DateFrom = t
	}

	if m.DateTo != nil {
		t, err := ptypes.Timestamp(m.DateTo)
		if err != nil {
			return nil, err
		}

		st.DateTo = t
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
	m.Type = decoded.Type

	m.CreatedAt, err = ptypes.TimestampProto(decoded.CreatedAt)
	if err != nil {
		return err
	}

	m.DateFrom, err = ptypes.TimestampProto(decoded.DateFrom)
	if err != nil {
		return err
	}

	m.DateTo, err = ptypes.TimestampProto(decoded.DateTo)
	if err != nil {
		return err
	}

	return nil
}
