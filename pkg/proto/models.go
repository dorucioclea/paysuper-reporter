package proto

type GeneratorPayload struct {
	Template *GeneratorTemplate `json:"template"`
	Options  *GeneratorOptions  `json:"options"`
	Data     interface{}        `json:"data"`
}

type GeneratorTemplate struct {
	ShortId string `json:"shortid,omitempty"`
	Name    string `json:"name,omitempty"`
	Recipe  string `json:"recipe,omitempty"`
	Content string `json:"content,omitempty"`
}

type GeneratorOptions struct {
	Timeout string `json:"timeout,omitempty"`
}
