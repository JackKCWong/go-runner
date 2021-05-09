package app

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"net/http"
	"os/exec"
	"sync"
)

type GoApp struct {
	*sync.Mutex
	Name   string `json:"name"`
	GitURL string `json:"gitUrl"`
	Status string `json:"status"`
	dir    string
	hc     *http.Client
	proc   *exec.Cmd
}

func (a *GoApp) Clone() error {
	a.Lock()
	defer a.Unlock()

	_, err := git.PlainClone(a.dir, false, &git.CloneOptions{
		URL:          a.GitURL,
		Depth:        1,
		SingleBranch: true,
	})

	if err != nil {
		return err
	}

	a.Status = "NEW"

	return nil
}

func (a *GoApp) Start() error {
	a.Lock()
	defer a.Unlock()

	buildCmd := exec.Command("go build -o " + a.Name)
	buildCmd.Path = a.dir

	if _, err := buildCmd.Output(); err != nil {
		a.Status = "ERR:BUILD"
		return err
	}

	sock, err := ioutil.TempFile(a.dir, "socket-")
	if err != nil {
		return err
	}

	runCmd := exec.Command(fmt.Sprintf("%s/%s -sock %s", a.dir, a.Name, sock.Name()))

	if err := runCmd.Start(); err != nil {
		a.Status = "ERR:RUN"
		return err
	}

	a.proc = runCmd
	a.Status = "STARTED"

	return nil
}

func (a *GoApp) Stop() error {
	return a.proc.Process.Kill()
}

func (a *GoApp) Handle(request *http.Request) (*http.Response, error) {
	return a.hc.Do(request)
}
