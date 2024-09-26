package cmd

import (
	"encoding/json"
	"fmt"
	"net"

	"mdepp/heu/internal/clients"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List entertainment configurations",
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(runListCmd())
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runListCmd() error {
	bridgeIP := net.ParseIP(viper.GetString("bridge_ip"))
	clipClient := clients.NewHueClipAPIClient(&clients.HueClipAPIClientConfig{
		BridgeIP:          bridgeIP,
		RootCertFilePath:  "hue-root.pem",
		HueApplicationKey: viper.GetString("hue_application_key"),
	})

	result, err := clipClient.ListEntertainmentConfigurations()
	if err != nil {
		return err
	}
	jsonResult, err := json.Marshal(result)
	if err != nil {
		return err
	}
	fmt.Println(string(jsonResult))

	return nil
}
