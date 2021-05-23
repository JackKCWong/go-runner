package app

import (
	testify "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

func TestGoRunner_Rehydrate(t *testing.T) {
	assert := testify.New(t)

	cwd, _ := os.Getwd()

	runner := &GoRunner{cwd: path.Join(cwd, "../..", "examples")}

	err := runner.Rehydrate()
	assert.Nil(err)

	goApp, err := runner.GetApp("hello-world")
	assert.Nil(err)

	time.Sleep(1 * time.Second)
	assert.Equal("STARTED", goApp.Status)
	assert.NotNil(goApp.proc)

	req, _ := http.NewRequest("GET", "http://nonehost/greeting", nil)
	resp, err := goApp.Handle(req)

	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(string(body), "hello world")
}
