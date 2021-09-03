package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/JackKCWong/go-runner/internal/util"
	"github.com/JackKCWong/go-runner/internal/web"
	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push the current git repo to remote and deploy it as an app to go-runner",
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Printf("failed to get verbose flag: %q\n", err)
			return err
		}

		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("failed to get current working dir: %q\n", err)
			return err
		}

		repo, err := git.PlainOpen(wd)
		if err != nil {
			fmt.Printf("failed to open current git repo: %q\n", err)
			return err
		}

		remote, err := repo.Remote("origin")
		if err != nil {
			fmt.Printf("failed to open current git repo: %q\n", err)
			return err
		}

		appName := path.Base(wd)
		gitURL := remote.Config().URLs[0]

		if verbose {
			fmt.Printf("verbose: pushing to remote origin: %s\n", gitURL)
		}

		sshAuth, err := util.GetGitAuth()
		if err != nil {
			return err
		}

		err = remote.Push(&git.PushOptions{Auth: sshAuth})
		if err != git.NoErrAlreadyUpToDate {
			fmt.Printf("failed to push current git repo: %q\n", err)
			return err
		}

		serverURL, err := cmd.Flags().GetString("server")
		if err != nil {
			fmt.Printf("failed to get server URL: %q\n", err)
			return err
		}

		endpoint := fmt.Sprintf("%s/api/%s", serverURL, appName)

		if verbose {
			fmt.Printf("verbose: deploying to %s... app=%s\n",
				endpoint, appName)
		}

		params := web.UpdateAppParams{
			App:    appName,
			Action: "deploy",
		}

		reqBody, err := json.Marshal(params)
		if err != nil {
			fmt.Printf("failed to create request params: %q\n", err)
			return err
		}

		req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(reqBody))
		if err != nil {
			fmt.Printf("failed to create request: %q\n", err)
			return err
		}

		resp, err := doREST(req)
		if err != nil {
			fmt.Printf("failed to complete request: %q\n", err)
			return err
		}

		fmt.Println(resp)

		return nil
	},
}
