package core

import (
	"context"
	"github.com/go-git/go-git/v5"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
)

type GoApp struct {
	_ struct{}
	sync.Mutex
	Name    string `json:"name"`
	GitURL  string `json:"gitUrl"`
	Status  string `json:"status"`
	AppDir  string `json:"appDir"`
	Cmd     string `json:"cmd"`
	LastErr error  `json:"lastError"`
	hc      *http.Client
	proc    *exec.Cmd
}

func NewGoApp(name string, appDir string) *GoApp {
	return &GoApp{Name: name, AppDir: appDir}
}

func (a *GoApp) purgeLocal() error {
	return os.RemoveAll(a.AppDir)
}

func (a *GoApp) Deploy(gitURL string) error {
	a.Lock()
	defer a.Unlock()

	err := a.purgeLocal()
	if err != nil {
		a.Status = "ERR:DEPLOY"
		a.LastErr = err
		return err
	}

	a.GitURL = gitURL

	_, err = git.PlainClone(a.AppDir, false, &git.CloneOptions{
		URL:          a.GitURL,
		Depth:        1,
		SingleBranch: true,
	})

	if err != nil {
		a.Status = "ERR:GITCLONE"
		a.LastErr = err
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
		a.LastErr = err
		return err
	}

	exePath := path.Join(a.AppDir, a.Name)
	sockPath := path.Join(a.AppDir, "sock")

	runCmd := exec.Command(exePath, "-sock", sockPath)
	runCmd.Dir = a.AppDir

	a.Cmd = runCmd.String()

	if err := runCmd.Start(); err != nil {
		a.Status = "ERR:START"
		a.LastErr = err
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

func (a *GoApp) Reload() error {
	a.Lock()
	defer a.Unlock()

	repo, err := git.PlainOpen(a.AppDir)
	if err != nil {
		return err
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return err
	}

	a.GitURL = remote.String()

	return nil
}

func (a *GoApp) Stop() error {
	return a.proc.Process.Kill()
}

func (a *GoApp) Handle(request *http.Request) (*http.Response, error) {
	req := request.Clone(context.TODO())
	req.Host = "sock"
	req.RequestURI = ""
	req.URL.Scheme = "http"
	req.URL.Host = "sock"
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+a.Name)
	return a.hc.Do(req)
}
