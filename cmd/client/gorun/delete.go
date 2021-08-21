package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete [appName]",
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"rm"},
	Short:   "Delete the app from go-runner",
	RunE: func(cmd *cobra.Command, args []string) error {
		var appName string
		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			fmt.Printf("failed to get verbose flag: %q\n", err)
			return err
		}

		if len(args) == 1 {
			appName = args[0]
		} else {
			if verbose {
				fmt.Printf("verbose: use basename as appName")
			}

			wd, err := os.Getwd()
			if err != nil {
				fmt.Printf("failed to get current working dir: %q\n", err)
				return err
			}

			appName = path.Base(wd)
		}

		serverURL, err := cmd.Flags().GetString("server")
		if err != nil {
			fmt.Printf("failed to get server URL: %q\n", err)
			return err
		}

		endpoint := fmt.Sprintf("%s/api/%s", serverURL, appName)
		req, err := http.NewRequest("DELETE", endpoint, nil)
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
