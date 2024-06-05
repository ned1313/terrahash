/*
Copyright © 2024 NAME HERE ned@nedinthecloud.com
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "terrahash",
	Short: "A CLI tool to generate and check hashes of Terraform modules",
	Long: `A CLI tool that can generate hashes of modules used by your Terraform configuration.
	Store the hashes, versions, and constraints in a local file. And check for module changes
	after initialization of Terraform.`,
}

var Source string

type moduleEntry struct {
	Key     string `json:"Key"`
	Source  string `json:"Source"`
	Version string `json:"Version"`
	Dir     string `json:"Dir"`
	Hash    string `json:"Hash"`
}

type modules struct {
	Modules []moduleEntry `json:"Modules"`
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terrahash.yaml)")
	rootCmd.PersistentFlags().StringVarP(&Source, "source", "s", "", "Source directory to read from. Defaults to current directory.")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
