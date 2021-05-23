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

func TestGoApp_Start(t *testing.T) {
	assert := testify.New(t)

	cwd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	goApp := NewGoApp("helloworld", path.Join(cwd, "../..", "examples", "apps", "hello-world"))

	err = goApp.Start()
	if err != nil {
		t.Error(err)
		return
	}

	defer goApp.Stop()

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
