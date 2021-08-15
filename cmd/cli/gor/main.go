package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "gor",
	Short: "Gor is a cli for go-runner to do CRUD for apps.",
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "base url to go-runner server.")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))

	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(deleteCmd)
}

func initConfig() {
	viper.SetEnvPrefix("go_runner")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
