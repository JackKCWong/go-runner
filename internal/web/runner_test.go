package web

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
	wd := path.Join(cwd, "../..", "examples")
	runner := NewWebServer(wd, ":8080")

	go runner.Start()

	time.Sleep(1 * time.Second)

	req, _ := http.NewRequest("GET", "http://localhost:8080/hello-world/greeting", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal("hello world", string(body))
}
