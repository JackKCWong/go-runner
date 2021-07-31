package core

import (
	"context"
	"encoding/json"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
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
	lastErr error
	hc      *http.Client
	proc    *exec.Cmd
}

func (a *GoApp) Purge() error {
	a.Lock()
	defer a.Unlock()

	a.Status = "DELETED"

	return os.RemoveAll(a.AppDir)
}

func (a *GoApp) Rebuild() error {
	a.Lock()
	defer a.Unlock()

	err := os.RemoveAll(a.AppDir)
	if err != nil {
		a.Status = "ERR:PURGE"
		a.lastErr = err
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		a.Status = "ERR:USERHOME"
		a.lastErr = err
		return err
	}

	sshKeyFile := path.Join(homeDir, ".ssh", "id_rsa")
	sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyFile, "")

	if err != nil {
		a.Status = "ERR:SSHKEY"
		a.lastErr = err
		return err
	}

	_, err = git.PlainClone(a.AppDir, false, &git.CloneOptions{
		URL:          a.GitURL,
		Depth:        1,
		SingleBranch: true,
		Auth:         sshAuth,
	})

	if err != nil {
		a.Status = "ERR:GITCLONE"
		a.lastErr = err
		return err
	}

	a.Status = "NEW"

	return nil
}

func (a *GoApp) Start() error {
	a.Lock()
	defer a.Unlock()

	//buildCmd := exec.Command("go", "build", "-o", a.Name)
	buildCmd := exec.Command("go", "build")
	buildCmd.Dir = a.AppDir

	if _, err := buildCmd.Output(); err != nil {
		a.Status = "ERR:BUILD"
		a.lastErr = err
		return err
	}

	exePath := path.Join(a.AppDir, a.Name)
	//exePath := a.Name
	sockPath := path.Join(a.AppDir, "sock")

	runCmd := exec.Command(exePath, "-sock", sockPath)
	runCmd.Dir, _ = os.Getwd()

	a.Cmd = runCmd.String()

	if err := runCmd.Start(); err != nil {
		a.Status = "ERR:START"
		a.lastErr = err
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

func (a *GoApp) Reattach() error {
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

	a.GitURL = remote.Config().URLs[0]

	return nil
}

func (a *GoApp) Stop() error {
	a.Lock()
	defer a.Unlock()

	if a.Status == "STARTED" {
		if err := a.proc.Process.Kill(); err != nil {
			return nil
		}

		a.Status = "STOPPED"
	}

	return nil
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

func (a *GoApp) Pull() {

}

func (a *GoApp) MarshalJSON() ([]byte, error) {
	var errMsg string
	if a.lastErr != nil {
		errMsg = a.lastErr.Error()
	}

	return json.Marshal(struct {
		Name    string `json:"name"`
		GitURL  string `json:"gitUrl"`
		Status  string `json:"status"`
		AppDir  string `json:"appDir"`
		Cmd     string `json:"cmd"`
		LastErr string `json:"lastError"`
	}{
		a.Name, a.GitURL, a.Status, a.AppDir, a.Cmd, errMsg,
	})
}
