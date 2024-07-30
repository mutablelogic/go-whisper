package schema

import (
	"encoding/json"
)

//////////////////////////////////////////////////////////////////////////////
// TYPES

type Model struct {
	Id      string `json:"id" writer:",width:28,wrap"`
	Object  string `json:"object,omitempty" writer:"-"`
	Path    string `json:"path,omitempty" writer:",width:40,wrap"`
	Created int64  `json:"created,omitempty"`
	OwnedBy string `json:"owned_by,omitempty"`
}

//////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (m *Model) String() string {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(data)
}
