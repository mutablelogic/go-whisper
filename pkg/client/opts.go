package client

// Request options
type opts struct {
	Language    string  `json:"language,omitempty"`
	Prompt      string  `json:"prompt,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
}

type Opt func(*opts) error

///////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS

func OptLanguage(language string) Opt {
	return func(o *opts) error {
		o.Language = language
		return nil
	}
}

func OptPrompt(prompt string) Opt {
	return func(o *opts) error {
		o.Prompt = prompt
		return nil
	}
}

func OptTemperature(t float32) Opt {
	return func(o *opts) error {
		o.Temperature = t
		return nil
	}
}
