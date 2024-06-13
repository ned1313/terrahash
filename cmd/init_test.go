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

	"github.com/stretchr/testify/assert"
)

func TestInitCmd(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Missing folder
	t.Run("MissingFolder", func(t *testing.T) {
		actual := new(bytes.Buffer)
		rootCmd.SetOut(actual)
		rootCmd.SetErr(actual)
		rootCmd.SetArgs([]string{"init", "--source", "missing"})
		rootCmd.Execute()

		expected := "Error: stat missing: no such file or directory\nUsage:\n  terrahash init [flags]\n\nFlags:\n  -h, --help   help for init\n\nGlobal Flags:\n  -s, --source string   Source directory to read from. Defaults to current directory.\n\n"

		assert.Equal(t, expected, actual.String())

	})

	// Test case 2: Terraform not initialized
	t.Run("TerraformNotInitialized", func(t *testing.T) {
		// Mock the setPath function to return the temporary directory
		actual := new(bytes.Buffer)
		rootCmd.SetOut(actual)
		rootCmd.SetErr(actual)
		rootCmd.SetArgs([]string{"init", "--source", tempDir})
		rootCmd.Execute()

		expected := ""

		assert.Equal(t, expected, actual.String())
	})

	// Test case 2: Mod lock file already exists
	t.Run("ModLockFileExists", func(t *testing.T) {
		// Create a mod lock file in the temporary directory
		modLockFile := tempDir + "/" + modFileName
		file, _ := os.Create(modLockFile)

		defer os.Remove(modLockFile)
		// TODO: Add assertions for the expected output
		file.WriteString(`{"Modules": {}}`)

		actual := new(bytes.Buffer)
		rootCmd.SetOut(actual)
		rootCmd.SetErr(actual)
		rootCmd.SetArgs([]string{"init", "--source", tempDir})
		rootCmd.Execute()

		expected := ""

		assert.Equal(t, expected, actual.String())
	})

	// Test case 3: No external modules found
	t.Run("NoExternalModulesFound", func(t *testing.T) {
		// Create a mod lock file in the temporary directory
		modLockFile := tempDir + "/" + modFileName
		file, _ := os.Create(modLockFile)

		defer os.Remove(modLockFile)
		file.WriteString(`{"Modules": {}}`)
		//Create an empty Terraform configuration
		terraformConfig := tempDir + "/main.tf"
		file, _ = os.Create(terraformConfig)
		defer os.Remove(terraformConfig)

		file.WriteString(`locals { testing = "test" }`)

		actual := new(bytes.Buffer)
		rootCmd.SetOut(actual)
		rootCmd.SetErr(actual)
		rootCmd.SetArgs([]string{"init", "--source", tempDir})
		rootCmd.Execute()

		expected := ""

		assert.Equal(t, expected, actual.String())
	})

	// Test case 4: External modules found
	t.Run("ExternalModulesFound", func(t *testing.T) {
		
		// TODO: Add assertions for the expected output
	})
}