/*
Copyright © 2024 the terrahash authors

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/
package cmd

import (
	"bytes"
	"os"
	"testing"
	"encoding/json"

	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestCheckCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Terraform not initialized
	t.Run("TerraformNotInitialized", func(t *testing.T) {
		// Create directory for this test
		testDir := tempDir + "/terraformnotinitialized"
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Mock the setPath function to return the temporary directory
		actual := new(bytes.Buffer)
		rootCmd.SetOut(actual)
		rootCmd.SetErr(actual)
		rootCmd.SetArgs([]string{"check", "--source", testDir})
		rootCmd.Execute()

		expected := ""
		assert.Equal(t, expected, actual.String())
	})

}

func TestCheck(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Non-matching modules by hash
	t.Run("NonMatchingModulesByHash", func(t *testing.T) {
		// Create directory for this test
		testDir := tempDir + "/nonmatchingmodulesbyhash"
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Add a configuration file to the directory
		configFile := testDir + "/main.tf"
		file, _ := os.Create(configFile)
		defer os.Remove(configFile)
		file.WriteString(testConfig)

		// Initialize Terraform
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: testDir,})

		terraform.Init(t, terraformOptions)

		// Create the mod lock file
		rootCmd.SetArgs([]string{"init", "--source", testDir})
		rootCmd.Execute()

		// Load the mod lock file
		lockedMods, err := processModFile(testDir + "/" + modFileName)

		if err != nil {
			t.Fatal(err)
		}

		// Edit the hash of the module
		lockedMods.Modules["vnet"] = moduleEntry{
			Key:     lockedMods.Modules["vnet"].Key,
			Source:  lockedMods.Modules["vnet"].Source,
			Version: lockedMods.Modules["vnet"].Version,
			Dir:     lockedMods.Modules["vnet"].Dir,
			Hash:    "newhash",
		}

		// Save the modified mod lock file
		//Prepare the json to look nice
		bytes, _ := json.MarshalIndent(lockedMods, "", "  ")

		// Create the mod lock file
		os.WriteFile(testDir + "/" + modFileName, bytes, os.ModePerm)

		// Run the check command
		checkErr := check(testDir + "/")

		assert.Contains(t, checkErr.Error(), "non matching or missing modules found in the configuration")


	})

	// Test case 2: Non-matching modules by version
	t.Run("NonMatchingModulesByVersion", func(t *testing.T) {
		// Create directory for this test
		testDir := tempDir + "/nonmatchingmodulesbyversion"
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Add a configuration file to the directory
		configFile := testDir + "/main.tf"
		file, _ := os.Create(configFile)
		defer os.Remove(configFile)
		file.WriteString(testConfig)

		// Initialize Terraform
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: testDir,})

		terraform.Init(t, terraformOptions)

		// Create the mod lock file
		rootCmd.SetArgs([]string{"init", "--source", testDir})
		rootCmd.Execute()

		// Load the mod lock file
		lockedMods, err := processModFile(testDir + "/" + modFileName)

		if err != nil {
			t.Fatal(err)
		}

		// Edit the hash of the module
		lockedMods.Modules["vnet"] = moduleEntry{
			Key:     lockedMods.Modules["vnet"].Key,
			Source:  lockedMods.Modules["vnet"].Source,
			Version: "0.0.0.0",
			Dir:     lockedMods.Modules["vnet"].Dir,
			Hash:    lockedMods.Modules["vnet"].Dir,
		}

		// Save the modified mod lock file
		//Prepare the json to look nice
		bytes, _ := json.MarshalIndent(lockedMods, "", "  ")

		// Create the mod lock file
		os.WriteFile(testDir + "/" + modFileName, bytes, os.ModePerm)

		// Run the check command
		checkErr := check(testDir + "/")

		assert.Contains(t, checkErr.Error(), "non matching or missing modules found in the configuration")

	})

	// Test case 3: New modules found
	t.Run("NewModulesFound", func(t *testing.T) {
		// Create directory for this test
		testDir := tempDir + "/newmodulesfound"
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Add a configuration file to the directory
		configFile := testDir + "/main.tf"
		file, _ := os.Create(configFile)
		defer os.Remove(configFile)
		file.WriteString(testConfig)

		// Initialize Terraform
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: testDir,})

		terraform.Init(t, terraformOptions)

		// Create an empty mod lock file
		modFile, _ := os.Create(testDir + "/" + modFileName)
		defer modFile.Close()
		modFile.WriteString(`{"Modules": {}}`)

		// Run the check command
		checkErr := check(testDir + "/")

		assert.Contains(t, checkErr.Error(), "non matching or missing modules found in the configuration")

	})

	// Test case 4: No changes required
	t.Run("NewChangesRequired", func(t *testing.T) {
		// Create directory for this test
		testDir := tempDir + "/nochanges"
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Add a configuration file to the directory
		configFile := testDir + "/main.tf"
		file, _ := os.Create(configFile)
		defer os.Remove(configFile)
		file.WriteString(testConfig)

		// Initialize Terraform
		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: testDir,})

		terraform.Init(t, terraformOptions)

		//Create the mod lock file
		rootCmd.SetArgs([]string{"init", "--source", testDir})
		rootCmd.Execute()

		// Run the check command
		checkErr := check(testDir + "/")

		assert.Nil(t, checkErr)

	})
}