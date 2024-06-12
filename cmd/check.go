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

	"github.com/spf13/cobra"
	"github.com/jedib0t/go-pretty/v6/table"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks if modules match the mod lock file",
	Long: `Checks to see if all the external modules used by the Terraform configuration match the 
	mod lock file. Will return an error if the hash for found modules do not match or if a module 
	is not found in the lock file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("check command called")
		// Get path value
		path, err := setPath(Source)
		if err != nil {
			return err
		}

		slog.Debug("check to see if terraform has been initialized")
		msg, init := terraformInitialized(path)

		if !init {
			slog.Warn(msg)
			slog.Warn("has terraform init been run?")
			return nil
		}

		return check(path)

	},
}

func check(path string) error {
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
			tw := table.NewWriter()
			tw.SetTitle("Non Matching Modules")
			tw.AppendHeader(table.Row{"Name","Version","Source"})
			for _,v := range noMatchHash.Modules {
				tw.AppendRow(table.Row{v.Key,v.Version,v.Source})
			}
			fmt.Println(tw.Render())
			fmt.Print("You can update these modules with the upgrade command.\n\n")
		}

		if len(notFoundMods.Modules) > 0 {
			tw := table.NewWriter()
			tw.SetTitle("Missing Modules")
			tw.AppendHeader(table.Row{"Name","Version","Source"})
			for _,v := range notFoundMods.Modules {
				tw.AppendRow(table.Row{v.Key,v.Version,v.Source})
			}
			fmt.Println(tw.Render())
			fmt.Print("You can add these modules with the upgrade command.\n\n")
		}

		if len(noMatchHash.Modules) > 0 || len(notFoundMods.Modules) > 0 {
			fmt.Printf("\nSummary: %v modules mising, %v non matching modules\n\n", len(notFoundMods.Modules), len(noMatchHash.Modules))
			return fmt.Errorf("non matching or missing modules found in the configuration")
		}

		fmt.Println("All modules match the mod lock file")
		return nil
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
