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

var registerCmd = &cobra.Command{
	Use:     "register",
	Aliases: []string{"reg", "pub"},
	Short:   "Push the the current git repo to remote origin and register it to go-runner",
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

		params := web.DeployAppParams{
			App:    path.Base(wd),
			GitUrl: remote.Config().URLs[0],
		}

		if verbose {
			fmt.Printf("verbose: pushing to remote origin: %s\n", params.GitUrl)
		}

		sshAuth, err := util.NewSshPubKeyAuth()
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

		endpoint := fmt.Sprintf("%s/api/apps", serverURL)

		if verbose {
			fmt.Printf("verbose: register to %s... app=%s, gitUrl=%s\n",
				endpoint, params.App, params.GitUrl)
		}

		reqPayload, err := json.Marshal(params)
		if err != nil {
			fmt.Printf("failed to create request params: %q\n", err)
			return err
		}

		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqPayload))
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
