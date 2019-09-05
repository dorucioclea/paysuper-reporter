package builder

import (
	"errors"
	"fmt"
	"github.com/paysuper/paysuper-reporter/pkg"
	errs "github.com/paysuper/paysuper-reporter/pkg/errors"
)

type Royalty DefaultHandler

func newRoyaltyHandler(h *Handler) BuildInterface {
	return &Royalty{Handler: h}
}

func (h *Royalty) Validate() error {
	if _, ok := h.report.Params[pkg.ParamsFieldId]; !ok {
		return errors.New(errs.ErrorParamIdNotFound.Message)
	}

	return nil
}

func (h *Royalty) Build() (interface{}, error) {
	royalty, err := h.royaltyReportRepository.GetById(fmt.Sprintf("%s", h.report.Params[pkg.ParamsFieldId]))

	if err != nil {
		return nil, err
	}

	return royalty, nil
}
