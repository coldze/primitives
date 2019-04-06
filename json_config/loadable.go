package json_config

import (
	"encoding/json"

	"github.com/coldze/primitives/custom_error"
)

type Loadable struct {
	DecodeTo func() interface{}
	PutTo    func(interface{}) custom_error.CustomError
}

func (l *Loadable) UnmarshalJSON(data []byte) error {
	cfg := l.DecodeTo()
	err := json.Unmarshal(data, cfg)
	if err != nil {
		return custom_error.MakeErrorf("Failed to unmarshal. Error: %v", err)
	}
	v, ok := cfg.(Validatable)
	if ok {
		cErr := v.Validate()
		if cErr != nil {
			return custom_error.NewErrorf(cErr, "Config validation failed.")
		}
	}
	l.PutTo(cfg)
	return nil
}
