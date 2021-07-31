package core

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestCanMarshalIntoJSON(t *testing.T) {
	expect := assert.New(t)
	goapp := &GoApp{
		Mutex:   sync.Mutex{},
		Name:    "hello-world",
		GitURL:  "git@test.git",
		Status:  "ERR:GITCLONE",
		AppDir:  "./",
		Cmd:     "./hello-world",
		lastErr: errors.New("testError"),
	}

	json, err := json.Marshal(goapp)
	expect.Nil(err)
	expect.Contains(string(json), "testError")
}
