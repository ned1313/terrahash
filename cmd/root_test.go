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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestTerraformInitialized(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Folder does not exist
	t.Run("FolderDoesNotExist", func(t *testing.T) {
		// Create a folder doesn't exist dir
		testDir := tempDir + "/doesntexist"
		err := os.Mkdir(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		actualMsg, actualBool := terraformInitialized(testDir)		

		expectedMsg, expectedBool := ".terraform directory was not found", false

		assert.Equal(t, expectedMsg, actualMsg)
		assert.Equal(t, expectedBool, actualBool)

	})

	// Test case 2: .terraform directory exists, mod lock file does not exist
	t.Run("FolderExists", func(t *testing.T) {
		// Create a .terraform directory in the temporary directory
		testDir := tempDir + "/folderexists"
		terraformDir := testDir + "/.terraform"
		err := os.MkdirAll(terraformDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		actualMsg, actualBool := terraformInitialized(testDir + "/")		

		expectedMsg, expectedBool := "modules.json file was not found", false

		assert.Equal(t, expectedMsg, actualMsg)
		assert.Equal(t, expectedBool, actualBool)
	})

	// Test case 3: .terraform directory exists, mod lock file exists
	t.Run("ModFileExists", func(t *testing.T) {
		// Create a .terraform/modules directory in the temporary directory
		testDir := tempDir + "/modfileexists"
		terraformDir := testDir + "/.terraform/modules"
		err := os.MkdirAll(terraformDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Create a modules.json file in the .terraform directory
		modulesFile := terraformDir + "/modules.json"
		fmt.Println(modulesFile)
		file, _ := os.Create(modulesFile)
		file.WriteString(`{"Modules": {}}`)

		actualMsg, actualBool := terraformInitialized(testDir + "/")		

		expectedMsg, expectedBool := "terraform has been initialized", true

		assert.Equal(t, expectedMsg, actualMsg)
		assert.Equal(t, expectedBool, actualBool)
	})
}

func TestProcessModules(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: modules.json file does not exist
	t.Run("ModulesFileDoesNotExist", func(t *testing.T) {
		testDir := tempDir + "/nomodfile"
		terraformDir := testDir + "/.terraform/modules"
		err := os.MkdirAll(terraformDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := processModules(testDir + "/")
		var expected modules
		expected.Modules = make(map[string]moduleEntry)

		assert.Equal(t, expected, actual)
		assert.NotNil(t, err)
	})

	// Test case 2: modules.json file exists, but can't be parsed
	t.Run("ModulesFileBad", func(t *testing.T) {
		// Create a .terraform/modules directory
		testDir := tempDir + "/modfilebad"
		terraformDir := testDir + "/.terraform/modules"
		err := os.MkdirAll(terraformDir, 0755)
		if err != nil {
			t.Fatal(err)
		}

		// Create a modules.json file in the .terraform/modules directory
		modulesFile := terraformDir + "/modules.json"
		file, _ := os.Create(modulesFile)

		// Write invalid JSON structure (expects a Module key with list value)
		file.WriteString(`{"Modules": {}}`)

		_, modErr := processModules(testDir + "/")

		assert.ErrorContains(t, modErr, "could not decode modules.json")
	})
	// Create a terraform configuration
	// Test case 3: modules.json file exists, can be parsed
	t.Run("ModulesFileGood", func(t *testing.T) {
		//Create Terraform configuration with external module
		testDir := tempDir + "/modfilegood"
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		terraformConfig := testDir + "/main.tf"
		file, _ := os.Create(terraformConfig)
		defer os.Remove(terraformConfig)

		file.WriteString(`
		provider "azurerm" {
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

		}`)

		terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
			TerraformDir: testDir,})

		terraform.Init(t, terraformOptions)

		_, modErr := processModules(testDir + "/")

		assert.Nil(t, modErr)
	})
}

func TestProcessModFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: lock file does not exist
	t.Run("LockFileDoesNotExist", func(t *testing.T) {
		testDir := tempDir + "/nolockfile"
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		actual, err := processModFile(testDir + "/" + modFileName)
		var expected modules

		assert.Equal(t, expected, actual)
		assert.NotNil(t, err)
	})

	// Test case 2: lock file exists, but can't be parsed
	t.Run("LockFileBad", func(t *testing.T) {
		// Create a lock file
		testDir := tempDir + "/lockfilebad"
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		lockFile := testDir + "/" + modFileName
		file, _ := os.Create(lockFile)

		// Write invalid JSON structure (expects a Modules key with map value)
		file.WriteString(`{"Modules": []}`)

		_, modErr := processModFile(lockFile)

		assert.ErrorContains(t, modErr, "could not decode lock file")
	})

	// Test case 3: lock file exists, can be parsed
	t.Run("LockFileGood", func(t *testing.T) {
		// Create a lock file
		testDir := tempDir + "/lockfilegood"
		err := os.MkdirAll(testDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
		lockFile := testDir + "/" + modFileName
		file, _ := os.Create(lockFile)

		// Write valid JSON structure
		file.WriteString(`{"Modules": {}}`)

		actual, modErr := processModFile(lockFile)
		var expected modules
		expected.Modules = make(map[string]moduleEntry)

		assert.Equal(t, expected, actual)
		assert.Nil(t, modErr)
	})
}

func TestSetPath (t *testing.T) {

	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: path is empty, should return current working directory
	t.Run("PathIsEmpty", func(t *testing.T) {
		actual, err := setPath("")
		expected, _ := os.Getwd()

		assert.Equal(t, expected + "/", actual)
		assert.Nil(t, err)
	})

	// Test case 2: path is not empty and directory does not exist
	t.Run("PathIsNotEmpty", func(t *testing.T) {
		testPath := tempDir + "/doesnotexist"
		actual, err := setPath(testPath)

		assert.Equal(t, testPath, actual)
		assert.NotNil(t, err)
	})

	// Test case 3: path is not empty and directory exists
	t.Run("PathIsNotEmptyAndExists", func(t *testing.T) {
		testPath := tempDir + "/exists"
		err := os.Mkdir(testPath, 0755)
		if err != nil {
			t.Fatal(err)
		}

		actual, err := setPath(testPath)

		assert.Equal(t, testPath + "/", actual)
		assert.Nil(t, err)
	})
}