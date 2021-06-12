package web

import (
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/core"
)

type (
	DeployAppParams struct {
		App    string `json:"app" validate:"required"`
		GitUrl string `json:"gitUrl" validate:"required"`
	}

	UpdateAppParams struct {
		App    string `json:"app" validate:"required"`
		Action string `json:"action" validate:"required"`
	}
)
type errStatus struct {
	*core.GoApp
	Error error
}

func (e errStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		App *core.GoApp `json:"app"`
		Err string      `json:"err"`
	}{e.GoApp, fmt.Sprintf("%s", e.Error)})
}
