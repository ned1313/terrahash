/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if modules match the mod lock file",
	Long: `Checks to see if all the external modules used by the Terraform configuration match the 
	mod lock file. Will return an error if the hash for found modules do not match or if a module 
	is not found in the lock file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("check command called")
		// Get path value
		var path string
		// Check to see if the .terraform directory exists
		slog.Debug("check to see if the .terraform directory exists")
		if Source != "" {
			path = Source
			slog.Info("working path set to source: " + Source)
		} else {
			pathCwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("unable to find the current working directory: %v", err)
			}
			slog.Info("working path set to current directory:" + pathCwd)
			path = pathCwd
		}

		// Make sure terraform is initialized
		if err := terraformInitialized(path); err != nil {
			return fmt.Errorf("terraform not initialized: %v", err)
		}

		// Load the sourced modules
		// Get the modules used by the configuration
		slog.Debug("get the modules used by the configuration")
		sourcedMods, err := processModules(path)
		if err != nil {
			return fmt.Errorf("error processing modules %v", err)
		}

		// Load the .terraform.modules.hcl file
		// If the file is not found throw an error
		lockedMods, err := processModFile(path + modFileName)
		if err != nil {
			return fmt.Errorf("error processing %v file: %v", modFileName, err)
		}

		// Compare the two files
		// For each entry in sourcedMods, check if an existing entry exists in lockedMods
		var notFoundMods, noMatchMods modules
		for _,s := range sourcedMods.Modules {
			var modFound bool = false
			for _,l := range lockedMods.Modules {
				if s.Key == l.Key {
					modFound = true
					if s.Hash != l.Hash {
						noMatchMods.Modules = append(noMatchMods.Modules, s)
					}
					break
				}
			}
			if !modFound {
				notFoundMods.Modules = append(notFoundMods.Modules, s)
			}
		}

		if len(noMatchMods.Modules) > 0 {
			fmt.Println("Non matching modules were found:")
			bytes, _ := json.MarshalIndent(noMatchMods.Modules, "", "  ")
			fmt.Println(string(bytes))
			fmt.Println("You may wish to update the module lock file using the upgrade command.")
		}

		if len(notFoundMods.Modules) > 0 {
			fmt.Println("The following modules were not found in the lock file:")
			bytes, _ := json.MarshalIndent(notFoundMods.Modules, "", "  ")
			fmt.Println(string(bytes))
			fmt.Println("You may wish to add these modules using the upgrade command.")
		}

		if len(noMatchMods.Modules) > 0 || len(notFoundMods.Modules) > 0 {
			return fmt.Errorf("non matching or missing modules found in the configuration")
		}

		slog.Info("all modules match the lock file")
		return nil

	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
