/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		// err := config.SetSubdir("client")
		// if err != nil {
		// 	panic(err)
		// }

		// // address is required
		// addr := cmd.Flag("addr").Value.String()
		// if len(addr) == 0 {
		// 	fmt.Println("Address is required (--addr)")
		// 	return
		// }

		// username, err := util.AskForString("Enter username")
		// if err != nil {
		// 	panic(err)
		// }

		// password, err := util.AskForPassword("Enter password")
		// if err != nil {
		// 	panic(err)
		// }

		// totpCode, err := util.AskForTotpCode(username)
		// if err != nil {
		// 	panic(err)
		// }

		// err = rpc.Login(addr, username, password, totpCode)
		// if err != nil {
		// 	panic(err)
		// }

	},
}

func init() {
	cliCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringP("addr", "a", "", "example-rmm.com:1234")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
