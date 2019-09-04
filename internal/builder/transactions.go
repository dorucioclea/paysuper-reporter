package builder

type Transactions DefaultHandler

func newTransactionsHandler(h *Handler) BuildInterface {
	return &Transactions{Handler: h}
}

func (h *Transactions) Validate() error {
	return nil
}

func (h *Transactions) Build() (interface{}, error) {
	return nil, nil
}
