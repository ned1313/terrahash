/*
Copyright © 2024 NAME HERE ned@nedinthecloud.com
*/
package cmd

import (
	"os"
	"log/slog"
	"fmt"
	"strings"
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/gosimple/hashdir"
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

const modFileName = ".terraform.module.hcl"

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error(err.Error())
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

func terraformInitialized(path string) error {
	// Check to see if Terraform has been initialized
	slog.Debug("checking to see if Terraform has been initialized")
	if !checkInit(path + ".terraform") {
		return fmt.Errorf("terraform has not been initialized")
	}

	//Check to see if the modules.json file exists
	slog.Debug("check to see if the modules.json file exists")
	if _, err := os.Stat(path + ".terraform/modules/modules.json"); err != nil {
		return fmt.Errorf("no modules found in .terraform/modules directory: %v", err)
	}

	return nil
}

func processModules(path string) (modules, error) {
	slog.Debug("get the modules used by the configuration")
	    var mods modules
		var sourcedMods modules

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
				slog.Info("skipping module: " + m.Key)
			} else {
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
				sourcedMods.Modules = append(sourcedMods.Modules, newMod)
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

		// Process the modules.json file
		slog.Debug("processing the mod lock file")

		err = json.NewDecoder(moduleFile).Decode(&mods)
		if err != nil {
			return mods, fmt.Errorf("could not decode modules.json: %v", err)
		}
		return mods, nil
}