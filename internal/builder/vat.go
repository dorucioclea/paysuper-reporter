package builder

import (
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type Vat DefaultHandler

func newVatHandler(h *Handler) BuildInterface {
	return &Vat{Handler: h}
}

func (h *Vat) Validate() error {
	if _, ok := h.report.Params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Vat) Build() (interface{}, error) {
	return map[string]interface{}{"name": "test", "email": "test@test.com"}, nil

	vat, err := h.vatReportRepository.GetById(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	return vat, nil
}
