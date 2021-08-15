package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [module name]",
	Short: "Initialize a example app server using unixsocket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := os.Stat("main.go"); err == nil {
			// file exists
			fmt.Fprintln(os.Stderr, "main.go already exists. Remove it and try again.")
			return err
		}

		if _, err := os.Stat("go.mod"); err == nil {
			// file exists
			fmt.Fprintln(os.Stderr, "go.mod already exists. Remove it and try again.")
			return err
		}

		mainFile, err := os.Create("main.go")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create main.go")
			return err
		}
		defer mainFile.Close()

		modFile, err := os.Create("go.mod")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create go.mod")
			return err
		}
		defer modFile.Close()

		mainContent, err := Asset("examples/server-unixsocket.go")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to load example content")
			return err
		}

		_, err = mainFile.WriteString(strings.Replace(string(mainContent), "package examples", "package main", 1))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to write main.go")
			return err
		}

		moduleName := "hello-world"
		if len(args) == 1 {
			moduleName = args[0]
		}

		_, err = modFile.WriteString(fmt.Sprintf("module %s\n\ngo 1.14\n", moduleName))
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to write go.mod")
			return err
		}

		return nil
	},
}
