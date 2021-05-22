package app

import (
	testify "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
)

func TestGoApp_Start(t *testing.T) {
	assert := testify.New(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	goApp := NewGoApp("helloworld", path.Join(cwd, "examples", "hello-world"))

	err = goApp.Start()
	if err != nil {
		t.Error(err)
		return
	}

	defer goApp.Stop()

	assert.Equal("STARTED", goApp.Status)
	assert.NotNil(goApp.proc)

	req, _ := http.NewRequest("GET", "/greeting", nil)
	resp, err := goApp.Handle(req)

	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Contains(body, "hello world")
}
