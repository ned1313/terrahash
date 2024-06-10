/*
Copyright © 2024 the terrahash authors

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"log/slog"

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

		// Compare the two files
		// For each entry in sourcedMods, check if an existing entry exists in lockedMods
		var notFoundMods, noMatchHash modules
		notFoundMods.Modules = make(map[string]moduleEntry)
		noMatchHash.Modules = make(map[string]moduleEntry)
		
		for k, s := range sourcedMods.Modules {
			// See if module is found in lockedMods
			l, ok := lockedMods.Modules[k]
			// If the module is found, check the Hash
			if ok {
				// If the hash or version doesn't match, add it to noMatchHash
				if s.Hash != l.Hash || s.Version != l.Version {
					slog.Debug("adding" + k + "to no match" )
				  noMatchHash.Modules[k] = s
			    }else{
					slog.Debug(k + "hashes and versions match")
				}
			}else{
				// If the entry isn't found, add it to notFoundMods
				slog.Debug("adding" + k + "to not found")
				notFoundMods.Modules[k] = s
			}
		}

		if len(noMatchHash.Modules) > 0 {
			fmt.Println("Non matching modules were found:")
			bytes, _ := json.MarshalIndent(noMatchHash.Modules, "", "  ")
			fmt.Println(string(bytes))
			fmt.Println("You may wish to update the module lock file using the upgrade command.")
		}

		if len(notFoundMods.Modules) > 0 {
			fmt.Println("The following modules were not found in the lock file:")
			bytes, _ := json.MarshalIndent(notFoundMods.Modules, "", "  ")
			fmt.Println(string(bytes))
			fmt.Println("You may wish to add these modules using the upgrade command.")
		}

		if len(noMatchHash.Modules) > 0 || len(notFoundMods.Modules) > 0 {
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
