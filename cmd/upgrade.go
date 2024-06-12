/*
Copyright © 2024 the terrahash authors

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var autoApprove bool

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade the lock file from the configuration",
	Long: `Upgrade will replace the entries in the mod lock file with the module references found
	in the Terraform configuration. Later this will be updated to target specific modules.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		slog.Debug("upgrade command called")

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

		//updated and added are used for table generation and summary
		var updateMods, updated, added, unchanged modules
		updateMods.Modules = make(map[string]moduleEntry)
		updated.Modules = make(map[string]moduleEntry)
		added.Modules = make(map[string]moduleEntry)
		unchanged.Modules = make(map[string]moduleEntry)

		//Cycle through each sourceMod
		// All source Mods will be added to the updateMods variable
		// Logging provides context on what is changing
		// TODO: Add support for single module upgrades
		// TODO: Log modules present in lock file, but missing from update
		for k, s := range sourcedMods.Modules {
			l, ok := lockedMods.Modules[k]
			if ok {
				if l.Hash != s.Hash || l.Version != s.Version {
					slog.Debug("updating hash or version for module " + k)
					updated.Modules[k] = s
				}else{
					slog.Debug("no change to module " + k)
					unchanged.Modules[k] = s
				}
			}else{
				slog.Debug("adding new module " + k)
				added.Modules[k] = s
			}
			updateMods.Modules[k] = s
		}

		if len(added.Modules) == 0 && len(updated.Modules) == 0 {
			fmt.Println("No changes detected for modules. Exiting...")
			return nil
		}
		
		tw := table.NewWriter()
		tw.AppendHeader(table.Row{"Name","Version","Change","Source"})
		for _,v := range updated.Modules {
			tw.AppendRow(table.Row{v.Key,v.Version,"Update",v.Source})
		}
		for _,v := range added.Modules {
			tw.AppendRow(table.Row{v.Key,v.Version,"Added",v.Source})
		}
		for _,v := range unchanged.Modules {
			tw.AppendRow(table.Row{v.Key,v.Version,"Unchanged",v.Source})
		}

		fmt.Println("The following changes were detected:")
		fmt.Println(tw.Render())
		fmt.Printf("\nSummary: %v to update, %v to add, %v unchanged\n", len(updated.Modules),len(added.Modules),len(unchanged.Modules))
		if !autoApprove {
		fmt.Print("Confirm changes by entering yes: ")
		var in = bufio.NewReader(os.Stdin)
		confirm := []string{"yes","y","Yes","Y"}
		resp, _ := in.ReadString('\n')
		if !(slices.Contains(confirm, strings.TrimSpace(resp))){
			slog.Debug("changes not accepted for mod lock file update")
			return fmt.Errorf("changes not accepted for mod lock file update")
		}
	}

		// Write out new mod lock file to path
		//Prepare the json to look nice
		bytes, _ := json.MarshalIndent(updateMods, "", "  ")

		// Create the mod lock file
		slog.Debug("writing modules out to file")
		os.WriteFile(path + modFileName, bytes, os.ModePerm)

		fmt.Println("Changes to mod lock file have been made successfully!")

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
	upgradeCmd.Flags().BoolVar(&autoApprove,"auto-approve",false,"Automatically approve mod lock file changes")
}
