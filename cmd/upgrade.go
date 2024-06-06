/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"encoding/json"

	"github.com/spf13/cobra"
)

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade the lock file from the configuration",
	Long: `Upgrade will replace the entries in the mod lock file with the module references found
	in the Terraform configuration. Later this will be updated to target specific modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Info("upgrade command called")
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
			slog.Info("working path set to current directory: " + pathCwd)
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

		var updateMods modules
		//Cycle through each sourceMod
		for _, s := range sourcedMods.Modules {
			var matchFound bool = false
			//Check if the matching module is in the lockedMods
			for _, l := range lockedMods.Modules {
				if l.Key == s.Key && l.Source == s.Source {
					matchFound = true
					if l.Hash != s.Hash {
						slog.Info("updating hash for module " + s.Key)
					}else{
						slog.Debug("no change to module " + s.Key)
					}
					break
				}
			}
			if !matchFound {
				slog.Info("adding new module " + s.Key)
			}
			updateMods.Modules = append(updateMods.Modules, s)
		}

		// Write out new mod lock file to path
		//Prepare the json to look nice
		bytes, _ := json.MarshalIndent(updateMods, "", "  ")

		// Create the mod lock file
		slog.Debug("writing modules out to file")
		os.WriteFile(path + modFileName, bytes, os.ModePerm)


		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
