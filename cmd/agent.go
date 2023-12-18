/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/rahn-it/svalin/config"
	"github.com/rahn-it/svalin/system/agent"

	"github.com/spf13/cobra"
)

// agentCmd represents the agent command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("agent called")

		profile, err := config.OpenProfile("default", "agent")
		if err != nil {
			panic(err)
		}

		config := profile.Config()
		config.BindFlags(cmd.Flags())

		err = agent.Init(profile)
		if err != nil {
			panic(err)
		}

		agent, err := agent.Connect(profile)
		if err != nil {
			panic(err)
		}

		err = agent.Run()
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.PersistentFlags().StringP("agent.address", "a", "", "example-rmm.com:1234")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// agentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// agentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
