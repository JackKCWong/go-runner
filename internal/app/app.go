package app

import (
	"context"
	"github.com/go-git/go-git/v5"
	"net"
	"net/http"
	"os/exec"
	"path"
	"sync"
)

type GoApp struct {
	_ struct{}
	sync.Mutex
	Name   string `json:"name"`
	GitURL string `json:"gitUrl"`
	Status string `json:"status"`
	AppDir string `json:"appDir"`
	hc     *http.Client
	proc   *exec.Cmd
}

func NewGoApp(name string, appDir string) *GoApp {
	return &GoApp{Name: name, AppDir: appDir}
}

func (a *GoApp) Clone(gitURL string) error {
	a.Lock()
	defer a.Unlock()

	a.GitURL = gitURL

	_, err := git.PlainClone(a.AppDir, false, &git.CloneOptions{
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

	buildCmd := exec.Command("go", "build", "-o", a.Name)
	buildCmd.Dir = a.AppDir

	if _, err := buildCmd.Output(); err != nil {
		a.Status = "ERR:BUILD"
		return err
	}

	exePath := path.Join(a.AppDir, a.Name)
	sockPath := path.Join(a.AppDir, "sock")

	runCmd := exec.Command(exePath, "-sock", sockPath)
	runCmd.Dir = a.AppDir

	if err := runCmd.Start(); err != nil {
		a.Status = "ERR:RUN"
		return err
	}

	a.proc = runCmd
	a.hc = &http.Client{
		Transport: &http.Transport{DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sockPath)
		}},
	}

	a.Status = "STARTED"

	return nil
}

func (a *GoApp) Stop() error {
	return a.proc.Process.Kill()
}

func (a *GoApp) Handle(request *http.Request) (*http.Response, error) {
	request.Host = "sock"
	return a.hc.Do(request)
}
