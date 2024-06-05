/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"strings"

	"encoding/json"

	"github.com/gosimple/hashdir"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a .terraform.module.hcl file if one doesn't already exist.",
	Long: `Init scans the current Terraform configuration and produces a .terraform.module.hcl
	file if one doesn't already exist. This command will error if a .terraform.module.hcl file
	is found or the Terraform configuration hasn't been initialized yet.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
		var path string
		// Check to see if the .terraform directory exists
		if Source != "" {
			path = Source
		} else {
			pathCwd, err := os.Getwd()
			if err != nil {
				fmt.Println("Unable to find the current working directory")
				os.Exit(1)
			}
			path = pathCwd
		}

		// Check to see if Terraform has been initialized
		if !checkInit(path + ".terraform") {
			fmt.Println("Terraform has not been initialized.")
			os.Exit(1)
		}

		//Check to see if the modules.json file exists
		if _, err := os.Stat(path + ".terraform/modules/modules.json"); err != nil {
			fmt.Println("No modules found in .terraform/modules directory.")
			os.Exit(0)
		}

		// Check to see if the .terraform.module.hcl file exists
		if _, err := os.Stat(path + ".terraform.modules.hcl"); err == nil {
			fmt.Println(".terraform.modules.hcl file already exists.")
			os.Exit(1)
		}
		// Get the modules used by the configuration
		moduleFile, err := os.Open(path + ".terraform/modules/modules.json")
		if err != nil {
			fmt.Println("Error opening modules.json file")
			os.Exit(1)
		}
		defer moduleFile.Close()

		// Get the version constraint for each module
		var mods modules

		err = json.NewDecoder(moduleFile).Decode(&mods)
		if err != nil {
			fmt.Println("Could not decode modules.json")
			os.Exit(1)
		}

		var sourcedMods modules
		// Remove module entries that are locally sourced
		for _, m := range mods.Modules {
			// All downloaded modules will reside in the .terraform/modules directory
			if strings.Split(m.Dir, "/")[0] != ".terraform" {
				fmt.Println("Skipping module", m.Key)
			} else {
				// Add a hash based on Dir contents
				hash, err := hashdir.Make(path+m.Dir, "sha256")
				if err != nil {
					fmt.Println("Could not create hash for", m.Key)
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Println("Hash generated:", hash)
				newMod := moduleEntry{
					Key:     m.Key,
					Dir:     m.Dir,
					Version: m.Version,
					Source:  m.Source,
					Hash:    hash,
				}
				sourcedMods.Modules = append(sourcedMods.Modules, newMod)
			}
		}

		fmt.Println(sourcedMods)

		if len(sourcedMods.Modules) == 0 {
			fmt.Println("No external modules found, exiting.")
			os.Exit(0)
		}

		//View the produced JSON
		bytes, _ := json.MarshalIndent(sourcedMods, "", "  ")
		fmt.Println(string(bytes))

		// Create the .terraform.module.hcl
		os.WriteFile(path + ".terraform.module.hcl", bytes, os.ModePerm)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkInit(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
