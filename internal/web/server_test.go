package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/core"
	testify "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
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

	runner := NewGoRunnerServer(tempDir)
	defer runner.Stop(context.Background())
	go runner.Start(":34567")

	assert.Eventuallyf(statusIsStarted("http://localhost:34567/api/health"), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	resp, err := http.DefaultClient.PostForm("http://localhost:34567/api/hello-world", url.Values{
		"app":    {"hello-world"},
		"gitUrl": {"git@github.com:JackKCWong/go-runner-hello-world.git"},
	})

	if err != nil {
		assert.FailNowf("failed to deploy app", "%q", err)
	}

	assert.Equal(http.StatusOK, resp.StatusCode)

	assert.Eventuallyf(hasApp("hello-world", "http://localhost:34567/api/health"), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
	assert.Eventuallyf(statusIsStarted("http://localhost:34567/api/hello-world"), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	resp, err = http.DefaultClient.Get("http://localhost:34567/hello-world/greeting")
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("hello world", string(body))

	// test delete app
	request, _ := http.NewRequest("DELETE", "http://localhost:34567/api/hello-world", nil)
	resp, err = http.DefaultClient.Do(request)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Eventuallyf(statusIsNotFound("http://localhost:34567/api/hello-world"), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
}

func hasApp(app, url string) func() bool {
	return func() bool {
		health, err := http.DefaultClient.Get(url)
		if err != nil {
			fmt.Printf("failed to get status: %q", err)
			return false
		}

		defer health.Body.Close()
		if health.StatusCode == 200 {
			body, err := ioutil.ReadAll(health.Body)
			if err != nil {
				fmt.Printf("failed to get status: %q", err)
				return false
			}

			status := struct {
				Apps []*core.GoApp
			}{}
			err = json.Unmarshal(body, &status)
			if err != nil {
				fmt.Printf("failed to unmarshal status: %q", err)
				return false
			}

			if len(status.Apps) == 1 {
				if status.Apps[0].Name == app {
					return true
				} else {
					return false
				}
			}
		}

		return false
	}
}

func statusIsNotFound(url string) func() bool {
	return func() bool {
		resp, err := http.DefaultClient.Get(url)
		if err != nil {
			return false
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return true
		}

		return false
	}
}

func statusIsStarted(url string) func() bool {
	return func() bool {
		health, err := http.DefaultClient.Get(url)
		if err != nil {
			fmt.Printf("failed to get status: %q", err)
			return false
		}

		defer health.Body.Close()
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
