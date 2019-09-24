package builder

import (
	"context"
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
	params, err := h.GetParams()

	if err != nil {
		return err
	}

	if _, ok := params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	if !bson.IsObjectIdHex(fmt.Sprintf("%s", params[pkg.ParamsFieldId])) {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Vat) Build() (interface{}, error) {
	params, _ := h.GetParams()
	vat, err := h.vatRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	return vat, nil
}

func (h *Vat) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
