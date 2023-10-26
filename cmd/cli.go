/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"rahnit-rmm/config"

	"github.com/spf13/cobra"
)

// cliCmd represents the cli command
var cliCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cli called")

		err := config.SetSubdir("client")
		if err != nil {
			panic(err)
		}

		// addr := "localhost:1234"

		// client, err := rpc.NewRpcClient(context.Background(), addr)

		// rpcCmd := &rpc.PingCmd{}

		// err = client.SendCommand(context.Background(), rpcCmd)
		// if err != nil {
		// 	panic(err)
		// }

		// err = client.Close(200, "OK")
		// if err != nil {
		// 	panic(err)
		// }

		// time.Sleep(time.Second)

	},
}

func init() {
	rootCmd.AddCommand(cliCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cliCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cliCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
