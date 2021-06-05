package web

type (
	DeployAppParams struct {
		App    string `json:"app" validate:"required"`
		GitUrl string `json:"gitUrl" validate:"required"`
	}
)
