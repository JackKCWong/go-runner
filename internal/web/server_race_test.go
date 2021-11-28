package web

import (
	"context"
	"fmt"
	testify "github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type testApp struct {
	reg    url.Values
	url    string
	expect string
}

func TestGoRunnerRaceCondition(t *testing.T) {
	assert := testify.New(t)

	tempDir, err := os.MkdirTemp(os.TempDir(), "go-runner-test")
	if err != nil {
		assert.FailNowf("failed to create temp dir", "%q", err)
	}

	fmt.Printf("starting at %s", tempDir)

	runner := NewGoRunnerServer(tempDir)
	defer runner.Stop(context.Background())

	err = runner.Bootsrap(":0")
	assert.Nil(err)

	go runner.Serve()

	// test health
	assert.Eventuallyf(statusIsStarted(runner.endpoint("/api/health")), 1*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	apps := []testApp{
		{
			url.Values{
				"app":    {"hello-world"},
				"gitUrl": {getExampleRepo("go-runner-hello-world")},
			},
			"/greeting",
			"hello world",
		},
		{
			url.Values{
				"app":    {"nihao-shijie"},
				"gitUrl": {getExampleRepo("go-runner-nihao-shijie")},
			},
			"/nihao",
			"nihao, 世界",
		},
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			rollIt(assert, apps[0], runner.endpoint)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			rollIt(assert, apps[1], runner.endpoint)
		}
	}()

	wg.Wait()
}

func rollIt(assert *testify.Assertions, app testApp, endpoint func(string) string) {
	resp, err := http.DefaultClient.PostForm(endpoint("/api/apps"), app.reg)

	if err != nil {
		assert.FailNowf("failed to deploy app", "%q", err)
	}

	assert.Equal(http.StatusOK, resp.StatusCode)

	appName := app.reg["app"][0]
	assert.Eventuallyf(hasApp(appName, endpoint("/api/health")), 5*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
	assert.Eventuallyf(statusIsStarted(endpoint("/api/"+appName)), 5*time.Second, 100*time.Millisecond, "timeout waiting for server to start")

	// test restart
	restartReq, _ := http.NewRequest("PUT", endpoint("/api/"+appName), strings.NewReader("app="+appName+"&action=restart"))
	restartReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err = http.DefaultClient.Do(restartReq)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Eventuallyf(hasApp(appName, endpoint("/api/health")), 5*time.Second, 100*time.Millisecond, "timeout waiting for app to deploy")
	assert.Eventuallyf(statusIsStarted(endpoint("/api/"+appName)), 5*time.Second, 100*time.Millisecond, "timeout waiting for app to start")

	// test app
	resp, err = http.DefaultClient.Get(endpoint("/" + appName + app.url))
	assert.Nil(err)

	body, err := ioutil.ReadAll(resp.Body)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal(app.expect, string(body))

	// test delete app
	deleteReq, _ := http.NewRequest("DELETE", endpoint("/api/"+appName), nil)
	resp, err = http.DefaultClient.Do(deleteReq)
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Eventuallyf(statusIsNotFound(endpoint("/api/"+appName)), 2*time.Second, 100*time.Millisecond, "timeout waiting for server to start")
}
