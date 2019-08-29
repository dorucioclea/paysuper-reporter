package builder

type Tax DefaultHandler

func newTaxHandler(h *Handler) BuildInterface {
	return &Tax{Handler: h}
}

func (h *Tax) Validate() error {
	return nil
}

func (h *Tax) Build() (interface{}, error) {
	return nil, nil
}
