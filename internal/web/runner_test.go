package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	testify "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestGoRunnerDeployApp(t *testing.T) {
	assert := testify.New(t)

	tempDir, err := os.MkdirTemp(os.TempDir(), "go-runner-test")
	if err != nil {
		assert.FailNowf("failed to create temp dir", "%q", err)
	}

	fmt.Printf("starting at %s", tempDir)

	runner := NewWebServer(tempDir)

	go func() {
		err := runner.Start(":34567")
		if err != nil {
			assert.FailNowf("failed to start server", "%q", err)
		}
	}()

	time.Sleep(1 * time.Second) // wait for go-runner start

	reqParams, _ := json.Marshal(DeployAppParams{App: "hello-world", GitUrl: "git@github.com:JackKCWong/go-runner-hello-world.git"})
	resp, err := http.DefaultClient.Post("http://localhost:34567/api/hello-world",
		"application/json", bytes.NewReader(reqParams))

	if err != nil {
		assert.FailNowf("failed to deploy app", "%q", err)
	}

	assert.Equal(http.StatusOK, resp.StatusCode)

	time.Sleep(1 * time.Second) // wait for app start

	resp, err = http.DefaultClient.Get("http://localhost:34567/hello-world/greeting")
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("hello world", string(body))
}
