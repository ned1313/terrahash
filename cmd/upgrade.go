/*
Copyright © 2024 the terrahash authors

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.

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
		// Get path value
		path, err := setPath(Source)
		if err != nil {
			return err
		}
		// Check to see if the .terraform directory exists
		slog.Debug("check to see if the .terraform directory exists")

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
		updateMods.Modules = make(map[string]moduleEntry)
		//Cycle through each sourceMod
		// All source Mods will be added to the updateMods variable
		// Logging provides context on what is changing
		// TODO: Add support for single module upgrades
		for k, s := range sourcedMods.Modules {
			l, ok := lockedMods.Modules[k]
			if ok {
				if l.Hash != s.Hash || l.Version != s.Version {
					slog.Info("updating hash or version for module " + k)
				}else{
					slog.Info("no change to module " + k)
				}
			}else{
				slog.Info("adding new module " + k)
			}
			updateMods.Modules[k] = s
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
