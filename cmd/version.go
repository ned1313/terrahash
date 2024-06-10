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
)

const coreVersion = "0.0.5"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "get the version of terrahash",
	Run: func(cmd *cobra.Command, args []string) {
		slog.Debug("version command called")
		fmt.Printf("terrahash version is %s\n", coreVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// upgradeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// upgradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
