package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/JackKCWong/go-runner/internal/util"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JackKCWong/go-runner/internal/core"
	testify "github.com/stretchr/testify/assert"
)

var tempDir string

func init() {
	var err error
	tempDir, err = os.MkdirTemp(os.TempDir(), "go-runner-test")
	if err != nil {
		panic(err)
	}

	fmt.Printf("starting at %s\n", tempDir)

	gitAuth, err := util.GetGitAuth()
	if err != nil {
		panic(err)
	}

	_, err = git.PlainClone(path.Join(tempDir, "go-runner-hello-world"), false,
		&git.CloneOptions{
			URL:   "git@github.com:JackKCWong/go-runner-hello-world.git",
			Depth: 1,
			Auth:  gitAuth,
		})

	if err != nil {
		panic(err)
	}

	_, err = git.PlainClone(path.Join(tempDir, "go-runner-nihao-shijie"), false,
		&git.CloneOptions{
			URL:   "git@github.com:JackKCWong/go-runner-nihao-shijie.git",
			Depth: 1,
			Auth:  gitAuth,
		})

	if err != nil {
		panic(err)
	}
}

func getExampleRepo(name string) string {
	dir, err := filepath.Abs(path.Join(tempDir, name))
	if err != nil {
		fmt.Printf("failed to get git repo dir: %q\n", err)
		panic(err)
	}

	return dir
}

func TestGoRunnerDeployApp(t *testing.T) {
	assert := testify.New(t)

	runner := NewGoRunnerServer(tempDir)
	defer runner.Stop(context.Background())

	err := runner.Bootsrap(":0")
	assert.Nil(err)

	go runner.Serve()

	// test health
	assert.Eventuallyf(statusIsStarted(runner.endpoint("/api/health")), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	resp, err := http.DefaultClient.PostForm(runner.endpoint("/api/apps"), url.Values{
		"app":    {"hello-world"},
		"gitUrl": {getExampleRepo("go-runner-hello-world")},
	})

	if err != nil {
		assert.FailNowf("failed to deploy app", "%q", err)
	}

	assert.Equal(http.StatusOK, resp.StatusCode)

	assert.Eventuallyf(hasApp("hello-world", runner.endpoint("/api/health")), 1*time.Second, 100*time.Millisecond, "timeout waiting for app to deploy")
	assert.Eventuallyf(statusIsStarted(runner.endpoint("/api/hello-world")), 1*time.Second, 100*time.Millisecond, "timeout waiting for app to start")

	// test restart
	restartReq, _ := http.NewRequest("PUT", runner.endpoint("/api/hello-world"), strings.NewReader("app=hello-world&action=restart"))
	restartReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err = http.DefaultClient.Do(restartReq)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Eventuallyf(hasApp("hello-world", runner.endpoint("/api/health")), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
	assert.Eventuallyf(statusIsStarted(runner.endpoint("/api/hello-world")), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	// test app
	resp, err = http.DefaultClient.Get(runner.endpoint("/hello-world/greeting"))
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("hello world", string(body))

	// test delete app
	deleteReq, _ := http.NewRequest("DELETE", runner.endpoint("/api/hello-world"), nil)
	resp, err = http.DefaultClient.Do(deleteReq)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Eventuallyf(statusIsNotFound(runner.endpoint("/api/hello-world")), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
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

			for _, a := range status.Apps {
				if a.Name == app {
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

		return resp.StatusCode == http.StatusNotFound
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

func (server *GoRunnerWebServer) endpoint(path string) string {
	return fmt.Sprintf("http://localhost:%d%s", server.port(), path)
}
