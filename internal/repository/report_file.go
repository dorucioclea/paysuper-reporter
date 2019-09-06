package repository

import (
	"github.com/globalsign/mgo/bson"
	database "github.com/paysuper/paysuper-database-mongo"
	"github.com/paysuper/paysuper-reporter/pkg/proto"
)

const (
	collectionReportFiles = "report_files"
)

type ReportFileRepositoryInterface interface {
	Insert(*proto.MgoReportFile) error
	Update(*proto.MgoReportFile) error
	GetById(string) (*proto.MgoReportFile, error)
}

func NewReportFileRepository(db *database.Source) ReportFileRepositoryInterface {
	s := &ReportFileRepository{db: db}
	return s
}

func (h *ReportFileRepository) Insert(rf *proto.MgoReportFile) error {
	if err := h.db.Collection(collectionReportFiles).Insert(rf); err != nil {
		return err
	}

	return nil
}

func (h *ReportFileRepository) Update(rf *proto.MgoReportFile) error {
	if err := h.db.Collection(collectionReportFiles).UpdateId(rf.Id, rf); err != nil {
		return err
	}

	return nil
}

func (h *ReportFileRepository) GetById(id string) (*proto.MgoReportFile, error) {
	var file *proto.MgoReportFile

	if err := h.db.Collection(collectionReportFiles).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&file); err != nil {
		return nil, err
	}

	return file, nil
}
