/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"rahnit-rmm/config"
	"rahnit-rmm/util"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
		_, err := config.GetCaCert()
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No root certificate found, generating one")

				encryptWith, err := util.AskForNewPassword("Enter password to encrypt the root certificate:")
				if err != nil {
					panic(err)
				}

				err = config.GenerateRootCert(encryptWith)
				if err != nil {
					fmt.Printf("Error generating root certificate: %v", err)
				}
			}
		} else {
			fmt.Println("Root certificate found, skipping CA generation")
		}

	},
}

func init() {
	cliCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
