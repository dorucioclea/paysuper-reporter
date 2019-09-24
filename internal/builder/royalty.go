package builder

import (
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type Royalty DefaultHandler

func newRoyaltyHandler(h *Handler) BuildInterface {
	return &Royalty{Handler: h}
}

func (h *Royalty) Validate() error {
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

func (h *Royalty) Build() (interface{}, error) {
	params, _ := h.GetParams()
	royalty, err := h.royaltyRepository.GetById(fmt.Sprintf("%s", params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	return royalty, nil
}

func (h *Royalty) PostProcess(ctx context.Context, id string, fileName string, retentionTime int) error {
	return nil
}
