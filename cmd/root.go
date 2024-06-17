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
	"os"
	"strings"

	"github.com/gosimple/hashdir"
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
	Modules map[string]moduleEntry `json:"Modules"`
}

type modulesFile struct {
	Modules []moduleEntry `json:"Modules"`
}

const modFileName = ".terraform.module.lock.hcl"

const testConfig = `provider "azurerm" {
			features{}
		}

		resource "azurerm_resource_group" "test" {
		  name = "terrahash-test"
		  location = "East US"
		}

		module "vnet" {
		  source  = "Azure/vnet/azurerm"
		  version = "4.1.0"

		    resource_group_name = azurerm_resource_group.test.name
		    use_for_each = true
		    vnet_location = azurerm_resource_group.test.location

		}`

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return err
	}
	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.terrahash.yaml)")
	rootCmd.PersistentFlags().StringVarP(&Source, "source", "s", "", "Source directory to read from. Defaults to current directory.")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

func terraformInitialized(path string) (string, bool) {
	// Check to see if Terraform has been initialized
	slog.Debug("checking to see if Terraform has been initialized")
	if !checkInit(path + ".terraform") {
		return ".terraform directory was not found", false
	}

	//Check to see if the modules.json file exists
	slog.Debug("check to see if the modules.json file exists")
	if _, err := os.Stat(path + ".terraform/modules/modules.json"); err != nil {
		return "modules.json file was not found", false
	}

	return "terraform has been initialized", true
}

func processModules(path string) (modules, error) {
	slog.Debug("get the modules used by the configuration")
	    var mods modulesFile
		var sourcedMods modules
		sourcedMods.Modules = make(map[string]moduleEntry)

		moduleFile, err := os.Open(path + ".terraform/modules/modules.json")
		if err != nil {
			return sourcedMods, fmt.Errorf("error opening modules.json file: %v", err)
		}
		defer moduleFile.Close()

		// Process the modules.json file
		slog.Debug("processing the modules.json file")

		err = json.NewDecoder(moduleFile).Decode(&mods)
		if err != nil {
			return sourcedMods, fmt.Errorf("could not decode modules.json: %v", err)
		}

		// Copy the external modules to sourcedMods
		slog.Debug("copying the external modules")
		
		for _, m := range mods.Modules {
			// All downloaded modules will reside in the .terraform/modules directory
			if strings.Split(m.Dir, "/")[0] != ".terraform" {
				slog.Debug("skipping module: " + m.Key)
			}else{
			// Add a hash based on Dir contents
				hash, err := hashdir.Make(path+m.Dir, "sha256")
				if err != nil {
					return sourcedMods, fmt.Errorf("could not create hash for %v: %v", m.Key, err)
				}
				slog.Debug("hash generated: " + hash)
				newMod := moduleEntry{
					Key:     m.Key,
					Dir:     m.Dir,
					Version: m.Version,
					Source:  m.Source,
					Hash:    hash,
				}
				// The root module shows up as an empty string
				if m.Key != "" {
					slog.Info("adding module: " + m.Key)
				  sourcedMods.Modules[m.Key] = newMod
				}
			}
		}

		slog.Debug("module processing complete:", "SourceMods", sourcedMods)
		return sourcedMods, nil
}

func processModFile(path string) (modules, error){
	var mods modules
	moduleFile, err := os.Open(path)
		if err != nil {
			return mods, fmt.Errorf("error opening mod lock file: %v", err)
		}
		defer moduleFile.Close()

		slog.Debug("processing the mod lock file")

		err = json.NewDecoder(moduleFile).Decode(&mods)
		if err != nil {
			return mods, fmt.Errorf("could not decode lock file: %v", err)
		}
		return mods, nil
}

func checkInit(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func setPath(path string) (string, error) {
	var setPath string
	if path == "" {
		setPath , err := os.Getwd()
		if err != nil {
			return path, fmt.Errorf("unable to find the current working directory: %v", err)
		}
		slog.Debug("working path set to current directory: " + setPath)
		return (setPath + "/"), nil
	} else {
		// test to make sure path exists
		if _, err := os.Stat(path); err != nil {
			return path, fmt.Errorf(err.Error())
		}
		setPath = path
		slog.Debug("working path set to source directory: " + setPath)
	}
	// If the path doesn't end with a '/' add it
	if strings.HasSuffix(strings.TrimSpace(setPath), "/") {
		slog.Debug("trailing slash found in " + setPath)
		return setPath, nil
	}
	slog.Debug("no trailing slash found in " + setPath)
	return (setPath + "/"), nil
}
