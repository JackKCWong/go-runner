package web

import (
	"bytes"
	"context"
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
	defer runner.Stop(context.Background())
	go runner.Start(":34567")

	assert.Eventuallyf(statusIsStarted("http://localhost:34567/api/health"), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	reqParams, _ := json.Marshal(DeployAppParams{App: "hello-world", GitUrl: "git@github.com:JackKCWong/go-runner-hello-world.git"})
	resp, err := http.DefaultClient.Post("http://localhost:34567/api/hello-world",
		"application/json", bytes.NewReader(reqParams))

	if err != nil {
		assert.FailNowf("failed to deploy app", "%q", err)
	}

	assert.Equal(http.StatusOK, resp.StatusCode)

	assert.Eventuallyf(statusIsStarted("http://localhost:34567/api/hello-world"), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
	resp, err = http.DefaultClient.Get("http://localhost:34567/hello-world/greeting")
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("hello world", string(body))

}

func statusIsStarted(url string) func() bool {
	return func() bool {
		health, err := http.DefaultClient.Get(url)
		if err != nil {
			fmt.Printf("failed to get status: %q", err)
			return false
		}

		if health.StatusCode == 200 {
			body, err := ioutil.ReadAll(health.Body)
			if err != nil {
				fmt.Printf("failed to get status: %q", err)
				return false
			}

			status := struct {
				Status string
			}{}
			err = json.Unmarshal(body, &status)
			if err != nil {
				fmt.Printf("failed to unmarshal status: %q", err)
				return false
			}

			if status.Status == "STARTED" {
				return true
			}
		}

		return false
	}
}
