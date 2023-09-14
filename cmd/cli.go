/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/spf13/cobra"
)

// cliCmd represents the cli command
var cliCmd = &cobra.Command{
	Use:   "cli",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cli called")

		target, err := net.ResolveUDPAddr("udp", "localhost:1234")
		if err != nil {
			panic(err)
		}

		fmt.Printf("\nconnecting to %s\n", target)

		port := 4321

		udpConn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: port})
		if err != nil {
			panic(err)
		}
		// ... error handling
		tr := quic.Transport{
			Conn: udpConn,
		}

		tlsConf := &tls.Config{}

		quicConf := &quic.Config{}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
		defer cancel()
		conn, err := tr.Dial(ctx, target, tlsConf, quicConf)

		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			panic(err)
		}

		header := map[string]interface{}{
			"type": "ping",
		}

		payload, err := json.Marshal(header)
		if err != nil {
			panic(err)
		}

		payload = append(payload, []byte("\n")...)

		_, err = stream.Write(payload)

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
