/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"rahnit-rmm/config"
	"rahnit-rmm/pki"
	"rahnit-rmm/rpc"
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
		config.SetSubdir("cli")

		var rootPassword []byte

		_, err := pki.GetCaCert()
		if err != nil {
			if errors.Is(err, pki.ErrNoCaCert) {
				fmt.Println("No root certificate found, generating one")

				rootUser, err := util.AskForString("Enter username for root:")
				if err != nil {
					panic(err)
				}

				rootPassword, err = util.AskForNewPassword("Enter password to encrypt the root certificate:")
				if err != nil {
					panic(err)
				}

				err = pki.InitCa(rootUser, rootPassword)
				if err != nil {
					fmt.Printf("error generating root certificate: %v", err)
				}
			} else {
				panic(err)
			}
		} else {
			fmt.Println("Root certificate found, skipping CA generation")
			rootPassword, err = util.AskForPassword("Enter password to decrypt the root certificate:")
			if err != nil {
				panic(err)
			}
		}

		pki.UnlockAsRoot(rootPassword)

		addr := "localhost:1234"

		client, err := rpc.NewRpcClient(context.Background(), addr)
		if err != nil {
			panic(err)
		}

		log.Println("Established client connection")

		rpcCmd, err := rpc.UploadCaCmd()
		if err != nil {
			panic(err)
		}

		log.Println("Preparing CA for Upload")

		err = client.SendCommand(context.Background(), rpcCmd)
		if err != nil {
			panic(err)
		}

		log.Println("CA Upload successful")

		err = client.Close(200, "OK")
		if err != nil {
			panic(err)
		}

		newUser, err := util.AskForString("Enter username for new user:")
		if err != nil {
			panic(err)
		}

		newPassword, err := util.AskForNewPassword("Enter password to encrypt the user certificate:")
		if err != nil {
			panic(err)
		}

		err = pki.CreateAndApplyCurrentUserCert(newUser, newPassword, rootPassword)
		if err != nil {
			panic(err)
		}
		pki.Unlock(newPassword)
		if err != nil {
			panic(err)
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
