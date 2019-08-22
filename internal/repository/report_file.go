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
	Insert(*proto.ReportFile) error
	Update(*proto.ReportFile) error
	GetById(string) (*proto.ReportFile, error)
	Delete(*proto.ReportFile) error
}

func NewReportFileRepository(db *database.Source) ReportFileRepositoryInterface {
	s := &ReportFileRepository{db: db}
	return s
}

func (h *ReportFileRepository) Insert(rf *proto.ReportFile) error {
	if err := h.db.Collection(collectionReportFiles).Insert(rf); err != nil {
		return err
	}

	return nil
}

func (h *ReportFileRepository) Update(rf *proto.ReportFile) error {
	if err := h.db.Collection(collectionReportFiles).UpdateId(bson.ObjectIdHex(rf.Id), rf); err != nil {
		return err
	}

	return nil
}

func (h *ReportFileRepository) Delete(rf *proto.ReportFile) error {
	if err := h.db.Collection(collectionReportFiles).RemoveId(bson.ObjectIdHex(rf.Id)); err != nil {
		return err
	}

	return nil
}

func (h *ReportFileRepository) GetById(id string) (*proto.ReportFile, error) {
	var file *proto.ReportFile

	if err := h.db.Collection(collectionReportFiles).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&file); err != nil {
		return nil, err
	}

	return file, nil
}
