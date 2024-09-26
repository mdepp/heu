package cmd

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"mdepp/heu/internal/clients"
	"mdepp/heu/internal/message"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type command struct {
	channelId uint8
	broadcast bool
	color     colorful.Color
}

type chanMessage struct {
	commands []command
	err      error
}

func parseCommands(text string) ([]command, error) {
	var commands []command
	for _, commandText := range strings.Split(text, ";") {
		commandTextChunks := strings.Split(commandText, " ")
		if len(commandTextChunks) == 1 {
			color, err := colorful.Hex(commandTextChunks[0])
			if err != nil {
				return nil, err
			}
			commands = append(commands, command{broadcast: true, color: color})
		} else if len(commandTextChunks) == 2 {
			channelId, err := strconv.ParseUint(commandTextChunks[0], 10, 8)
			if err != nil {
				return nil, err
			}
			color, err := colorful.Hex(commandTextChunks[1])
			if err != nil {
				return nil, err
			}
			commands = append(commands, command{channelId: uint8(channelId), color: color})
		}
	}
	return commands, nil
}

func getEntertainmentConfig(client *clients.HueClipAPIClient, id string) (*clients.EntertainmentConfiguration, error) {
	result, err := client.ListEntertainmentConfigurations()
	if err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, errors.New("no entertainment configurations exist")
	}
	if id == "" {
		return &result.Data[0], nil
	}
	for _, config := range result.Data {
		if config.ID == id {
			return &config, nil
		}
	}
	return nil, fmt.Errorf("unable to find entertainment configuration matching %s", id)
}

var (
	cmdFilePath string
	frequency   int
)

var streamCmd = &cobra.Command{
	Use:   "stream [CONFIG_ID]",
	Short: "Stream colour commands",
	Long: `Initiates an Entertainment API streaming session and streams colour commands
from standard input to the session.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var entertainmentConfigId string
		if len(args) > 0 {
			entertainmentConfigId = args[0]
		}
		cobra.CheckErr(runStreamCmd(entertainmentConfigId))
	},
}

func init() {
	rootCmd.AddCommand(streamCmd)

	streamCmd.PersistentFlags().StringVar(&cmdFilePath, "file", "", "file from which to read commands (defaults to standard input)")
	streamCmd.PersistentFlags().IntVar(&frequency, "frequency", 60, "frequency in hz at which to send entertainment API messages over UDP")
}

func runStreamCmd(entertainmentConfigId string) error {
	cmdFile := os.Stdin
	if cmdFilePath != "" {
		if cmdFile_, err := os.Open(cmdFilePath); err != nil {
			return err
		} else {
			cmdFile = cmdFile_
		}
	} else {
		slog.Info("Enter colour commands now (ctrl+d to exit)")
	}

	bridgeIP := net.ParseIP(viper.GetString("bridge_ip"))
	dtlsClientKey, err := hex.DecodeString(viper.GetString("dtls_client_key"))
	if err != nil {
		return err
	}

	clipClient := clients.NewHueClipAPIClient(&clients.HueClipAPIClientConfig{
		BridgeIP:          bridgeIP,
		RootCertFilePath:  "hue-root.pem",
		HueApplicationKey: viper.GetString("hue_application_key"),
	})

	entertainmentClient, err := clients.NewHueEntertainmentAPIClient(&clients.HueEntertainmentAPIClientConfig{
		BridgeIP:          bridgeIP,
		HueApplicationKey: viper.GetString("hue_application_key"),
		DTLSClientKey:     dtlsClientKey,
	})
	if err != nil {
		return err
	}
	defer entertainmentClient.Close()

	entertainmentConfig, err := getEntertainmentConfig(clipClient, entertainmentConfigId)
	if err != nil {
		return err
	}

	channelIds := []uint8{}
	for _, channel := range entertainmentConfig.Channels {
		channelIds = append(channelIds, channel.ChannelID)
	}

	if err = clipClient.StartEntertainmentSession(entertainmentConfig.ID); err != nil {
		return err
	}
	defer clipClient.StopEntertainmentSession(entertainmentConfig.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err = entertainmentClient.Handshake(ctx); err != nil {
		return err
	}

	commandChan := make(chan chanMessage)

	go func() {
		reader := bufio.NewReader(cmdFile)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				commandChan <- chanMessage{err: err}
				return
			}
			commands, err := parseCommands(line)
			if err != nil {
				commandChan <- chanMessage{err: err}
				return
			}
			commandChan <- chanMessage{commands: commands}

		}
		slog.Info("Exiting cleanly...")
		commandChan <- chanMessage{}
	}()

	ticker := time.NewTicker(time.Second / time.Duration(frequency))
	defer ticker.Stop()

	builder := message.NewBuilder(message.COLOR_SPACE_RGB).WritePreamble([]byte(entertainmentConfig.ID))
	message := builder.Build()

	quitting := false
	var result error

	for {
		select {
		case <-ticker.C:
			_, err = entertainmentClient.Write(message)
			if err != nil {
				return err
			}
			if quitting {
				return result
			}
		case chanMessage := <-commandChan:
			if len(chanMessage.commands) > 0 {
				builder.ResetBody()
				for _, userMessage := range chanMessage.commands {
					if userMessage.broadcast {
						for _, channelId := range channelIds {
							builder.WriteChannelColor(channelId, userMessage.color)
						}
					} else {
						builder.WriteChannelColor(userMessage.channelId, userMessage.color)
					}
				}
				message = builder.Build()
			} else {
				quitting = true
				result = chanMessage.err
			}
		}
	}
}
