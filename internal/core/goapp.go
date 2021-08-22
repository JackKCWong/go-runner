package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-cmd/cmd"
	"github.com/go-git/go-git/v5"
	"github.com/smallnest/ringbuffer"
)

type GoApp struct {
	_ struct{}
	sync.Mutex
	Name        string
	GitURL      string
	gitCommit   string
	Status      string
	AppDir      string
	lastErr     error
	buildStatus cmd.Status
	proc        *cmd.Cmd
	proxy       *httputil.ReverseProxy
	stdout      *ringbuffer.RingBuffer
	stderr      *ringbuffer.RingBuffer
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

	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	a.Status = "ERR:USERHOME"
	// 	a.lastErr = err
	// 	return err
	// }

	// sshKeyFile := path.Join(homeDir, ".ssh", "id_rsa")
	// sshAuth, err := ssh.NewPublicKeysFromFile("git", sshKeyFile, "")

	// if err != nil {
	// 	a.Status = "ERR:SSHKEY"
	// 	a.lastErr = err
	// 	return err
	// }

	repo, err := git.PlainClone(a.AppDir, false, &git.CloneOptions{
		URL:          a.GitURL,
		Depth:        1,
		SingleBranch: true,
		// Auth:         sshAuth,
	})

	if err != nil {
		a.Status = "ERR:GITCLONE"
		a.lastErr = err
		return err
	}

	err = a.attach(repo)
	if err != nil {
		a.Status = "ERR:GITOPEN"
		a.lastErr = err
		return err
	}

	a.Status = "NEW"

	return nil
}

func (a *GoApp) Start() error {
	a.Lock()
	defer a.Unlock()

	// buildCmd := exec.Command("go", "build", "-o", a.Name)
	buildCmd := cmd.NewCmd("go", "build", "-o", a.Name)
	buildCmd.Dir = a.AppDir

	a.buildStatus = <-buildCmd.Start()

	if a.buildStatus.Error != nil {
		a.Status = "ERR:BUILD"
		a.lastErr = a.buildStatus.Error
		return a.buildStatus.Error
	}

	exePath := path.Join(a.AppDir, a.Name)
	//exePath := a.Name
	sockPath := path.Join(a.AppDir, "sock")

	runCmd := cmd.NewCmdOptions(cmd.Options{
		Buffered:  false,
		Streaming: true,
	}, exePath, "-unixsock", sockPath)
	runCmd.Dir, _ = os.Getwd()

	a.stdout = ringbuffer.New(1024 * 100) // 100 kb
	go func(buf *ringbuffer.RingBuffer) {
		for line := range runCmd.Stdout {
			// consume stdout to avoid blocking
			_, _ = buf.WriteString(line)
		}
	}(a.stdout)

	a.stderr = ringbuffer.New(1024 * 100) // 100 kb
	go func(buf *ringbuffer.RingBuffer) {
		for line := range runCmd.Stderr {
			// consume stderr to avoid blocking
			_, _ = buf.WriteString(line)
		}
	}(a.stderr)

	runCmd.Start()
	<-time.After(100 * time.Millisecond) // give a little time for PID to be ready

	a.proc = runCmd
	targetURL, err := url.Parse("http://sock")
	if err != nil {
		return err
	}

	a.proxy = httputil.NewSingleHostReverseProxy(targetURL)
	a.proxy.Transport = &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", sockPath)
		}}

	a.Status = "STARTED"

	return nil
}

func (a *GoApp) attach(repo *git.Repository) error {
	head, err := repo.Head()
	if err != nil {
		a.Status = "ERR:GITLOG"
		a.lastErr = err
		return err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		a.Status = "ERR:GITLOG"
		a.lastErr = err
		return err
	}

	hash := head.Hash().String()[0:7]
	a.gitCommit = fmt.Sprintf("%s %s @ %s by %s",
		hash, strings.TrimRight(commit.Message, "\n"),
		commit.Author.String(), commit.Author.When.String())

	return nil
}

func (a *GoApp) Reattach() error {
	a.Lock()
	defer a.Unlock()

	repo, err := git.PlainOpen(a.AppDir)
	if err != nil {
		return err
	}

	return a.attach(repo)
}

func (a *GoApp) Stop() error {
	a.Lock()
	defer a.Unlock()

	if a.Status == "STARTED" {
		if err := a.proc.Stop(); err != nil {
			return err
		}

		a.Status = "STOPPED"
		return nil
	}

	return errors.New("app not started: status=" + a.Status)
}

func (a *GoApp) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, "/"+a.Name)
	a.proxy.ServeHTTP(rw, req)
}

func (a *GoApp) Pull() {

}

func (a *GoApp) MarshalJSON() ([]byte, error) {
	a.Lock()
	defer a.Unlock()

	var errMsg string
	if a.lastErr != nil {
		errMsg = a.lastErr.Error()
	}

	var status cmd.Status = cmd.Status{
		PID:  -1,
		Exit: -1,
	}

	if a.proc != nil {
		status = a.proc.Status()
	}

	return json.Marshal(struct {
		Name      string `json:"name"`
		GitURL    string `json:"gitUrl"`
		GitCommit string `json:"gitCommit"`
		Status    string `json:"status"`
		AppDir    string `json:"appDir"`
		LastErr   string `json:"lastError"`
		PID       int    `json:"pid"`
		Exit      int    `json:"exit"`
	}{
		a.Name, a.GitURL, a.gitCommit, a.Status, a.AppDir, errMsg,
		status.PID, status.Exit,
	})
}
