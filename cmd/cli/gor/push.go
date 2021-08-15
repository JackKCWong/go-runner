package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/JackKCWong/go-runner/internal/web"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current git repo to remote and deploy it as an app to go-runner",
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("failed to get current working dir: %q\n", err)
			return
		}

		repo, err := git.PlainOpen(wd)
		if err != nil {
			fmt.Printf("failed to open current git repo: %q\n", err)
			return
		}

		remote, err := repo.Remote("origin")
		if err != nil {
			fmt.Printf("failed to open current git repo: %q\n", err)
			return
		}

		err = remote.Push(&git.PushOptions{})
		if err != git.NoErrAlreadyUpToDate {
			fmt.Printf("failed to push current git repo: %q\n", err)
			return
		}

		serverURL, err := cmd.Flags().GetString("server")
		if err != nil {
			fmt.Printf("failed to get server URL: %q\n", err)
			return
		}

		params := web.DeployAppParams{
			App:    path.Base(wd),
			GitUrl: remote.Config().URLs[0],
		}

		endpoint := fmt.Sprintf("%s/api/apps", serverURL)

		if verbose, err := cmd.Flags().GetBool("verbose"); verbose && err == nil {
			fmt.Printf("verbose: pushing to %s... app=%s, gitUrl=%s\n",
				endpoint, params.App, params.GitUrl)
		}

		reqBody, err := json.Marshal(params)
		if err != nil {
			fmt.Printf("failed to create request params: %q\n", err)
			return
		}

		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Printf("failed to create http request: %q\n", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		req = req.WithContext(ctx)
		req.Header.Add("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("failed to get http response: %q\n", err)
			return
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("failed to read http response: %q\n", err)
			return
		}

		var prettyJSON bytes.Buffer
		json.Indent(&prettyJSON, body, "", "  ")

		fmt.Println(prettyJSON.String())
	},
}
